// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"fmt"
	"image"
	"runtime"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/internal/cache"
	"poly.red/internal/imageutil"
	"poly.red/material"
	"poly.red/math"
	"poly.red/shader"

	"poly.red/internal/profiling"
	"poly.red/internal/sched"
)

// 1 second / 60fps = 16.6666 ms/frame
// 1 second / 30fps = 33.3333 ms/frame

// Renderer is a hybrid software renderer that implements
// rasterization and ray tracing.
type Renderer struct {
	// rendering options
	cfg *option

	// scheduling, use for hard interruption.
	running uint32 // atomic
	stop    uint32 // atomic

	// rendering caches
	sched      *sched.Pool
	bufcur     int
	buflen     int
	bufs       []*buffer.FragmentBuffer
	shadowBufs []shadowInfo
	outBuf     *image.RGBA
}

// NewRenderer creates a new renderer.
//
// The returned renderer implements a rasterization rendering pipeline.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{ // default settings
		buflen: 2, // use 2 by default.
		bufs:   nil,
		cfg: &option{
			Width:     800,
			Height:    600,
			MSAA:      1,
			ShadowMap: false,
			Debug:     false,
			Scene:     nil,
			Workers:   runtime.NumCPU(),
			BatchSize: 32, // heuristic
			Format:    buffer.PixelFormatRGBA,
		},
	}
	for _, opt := range opts {
		opt(r.cfg)
	}

	r.bufs = make([]*buffer.FragmentBuffer, r.buflen)
	r.resetBufs()

	r.sched = sched.New(sched.Workers(r.cfg.Workers))
	runtime.SetFinalizer(r, func(r *Renderer) {
		r.sched.Release()
	})

	// initialize shadow maps
	if r.cfg.Scene != nil && r.cfg.ShadowMap {
		r.initShadowMaps()
		r.bufs[0].ClearFragment()
		r.bufs[0].ClearColor()
	}

	return r
}

// resetBuffers assign new buffers to the caches window buffers (w.bufs)
// Note: with Metal, we always use RGBA pixel format.
func (r *Renderer) resetBufs() {
	for i := 0; i < r.buflen; i++ {
		r.bufs[i] = buffer.NewBuffer(image.Rect(0, 0, r.cfg.Width*r.cfg.MSAA, r.cfg.Height*r.cfg.MSAA),
			buffer.Format(r.cfg.Format))
	}
}

// Render renders a scene.
func (r *Renderer) Render() *image.RGBA {
	if r.cfg.Debug {
		runtime.GOMAXPROCS(r.cfg.Workers)
		fmt.Printf("rendering under GOMAXPROCS=%v\n", r.cfg.Workers)
		total := profiling.Timed("entire rendering")
		defer total()
	}

	// record running
	r.startRunning()
	defer r.stopRunning()

	// reset buffers
	r.bufs[0].ClearColor()
	if r.shouldStop() {
		return r.outBuf
	}

	// decide if need shadow passes
	if r.cfg.ShadowMap {
		for i := 0; i < len(r.shadowBufs); i++ {
			r.passShadows(i)
			if r.shouldStop() {
				return r.outBuf
			}
		}
		r.bufs[0].ClearColor()
	}

	r.passForward()
	if r.shouldStop() {
		return r.outBuf
	}

	r.bufs[0].ClearColor()
	r.passDeferred()
	if r.shouldStop() {
		return r.outBuf
	}

	r.passAntialiasing()
	return r.outBuf
}

func (r *Renderer) CurrBuffer() *buffer.FragmentBuffer {
	return r.bufs[r.bufcur]
}

func (r *Renderer) NextBuffer() *buffer.FragmentBuffer {
	r.bufcur = (r.bufcur + 1) % r.buflen
	r.bufs[r.bufcur].Clear()
	return r.bufs[r.bufcur]
}

func (r *Renderer) passForward() {
	if r.cfg.Debug {
		done := profiling.Timed("forward pass (world)")
		defer done()
	}

	sum := 0
	r.cfg.Scene.IterMeshes(func(m mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool {
		sum += len(m.Triangles())
		r.sched.Add(len(m.Triangles()))
		return true
	})

	r.cfg.Scene.IterMeshes(func(m mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool {
		mesh := m.(*mesh.TriangleMesh)
		mvp := &shader.MVP{
			Model: mesh.ModelMatrix(),
			View:  r.cfg.Camera.ViewMatrix(),
			Proj:  r.cfg.Camera.ProjMatrix(),
			Viewport: math.ViewportMatrix(
				float32(r.bufs[0].Bounds().Dx()),
				float32(r.bufs[0].Bounds().Dy()),
			),
			Normal: mesh.ModelMatrix().Inv().T(),
		}
		mvp.ViewInv = mvp.View.Inv()
		mvp.ProjInv = mvp.Proj.Inv()
		mvp.ViewportInv = mvp.Viewport.Inv()

		tris := mesh.Triangles()
		for i := range tris {
			t := tris[i]
			r.sched.Run(func() {
				if !t.IsValid() {
					return
				}
				r.draw(mvp, t)
			})
		}
		return true
	})
	r.sched.Wait()
}

func (r *Renderer) passDeferred() {
	if r.cfg.Debug {
		done := profiling.Timed("deferred pass (shading)")
		defer done()
	}
	matView := r.cfg.Camera.ViewMatrix()
	matViewInv := matView.Inv()
	matProj := r.cfg.Camera.ProjMatrix()
	matProjInv := matProj.Inv()
	matVP := math.ViewportMatrix(float32(r.bufs[0].Bounds().Dx()), float32(r.bufs[0].Bounds().Dy()))
	matVPInv := matVP.Inv()
	matScreenToWorld := matViewInv.MulM(matProjInv).MulM(matVPInv)
	uniforms := &shader.MVP{
		View:            matView,
		ViewInv:         matViewInv,
		Proj:            matProj,
		ProjInv:         matProjInv,
		Viewport:        matVP,
		ViewportToWorld: matScreenToWorld,
	}

	r.DrawFragments(r.bufs[0], func(frag *primitive.Fragment) color.RGBA {
		return r.shade(frag, uniforms)
	})
}

func (r *Renderer) shade(frag *primitive.Fragment, uniforms *shader.MVP) color.RGBA {
	info := r.bufs[0].UnsafeGet(frag.X, frag.Y)
	if !info.Ok {
		return r.cfg.Background
	}

	col := info.Col
	mat := cache.Get[*material.BlinnPhong](info.MaterialID)
	if mat != nil {
		lightSources, lightEnv := r.cfg.Scene.Lights()
		col = mat.FragmentShader(
			info,
			r.cfg.Camera.Position(), lightSources, lightEnv)

		if r.cfg.ShadowMap && mat.ReceiveShadow {
			visibles := float32(0.0)
			ns := len(r.shadowBufs)
			for i := 0; i < ns; i++ {
				visible := r.shadingVisibility(i, info, uniforms)
				if visible {
					visibles++
				}
			}
			w := math.Pow(0.5, visibles)
			r := uint8(float32(col.R) * w)
			g := uint8(float32(col.G) * w)
			b := uint8(float32(col.B) * w)
			col = color.RGBA{r, g, b, col.A}
		}
	}

	// FIXME: why it has to be frag?
	frag.Col = col
	return material.AmbientOcclusionShade(r.bufs[0], frag)
}

func (r *Renderer) passAntialiasing() {
	if r.cfg.Debug {
		done := profiling.Timed("antialiasing")
		defer done()
	}

	// converts color from linear to sRGB space.
	if r.cfg.GammaCorrect {
		r.DrawFragments(r.bufs[0], shader.GammaCorrection)
	}
	r.outBuf = imageutil.Resize(r.cfg.Width, r.cfg.Height, r.bufs[0].Image())
}

func (r *Renderer) draw(
	mvp *shader.MVP,
	t *primitive.Triangle) {
	var t1, t2, t3 *primitive.Vertex
	t1 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V1.Pos),
		Col: t.V1.Col,
		UV:  t.V1.UV,
		Nor: t.V1.Nor.Apply(mvp.Normal),
	}
	t2 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V2.Pos),
		Col: t.V2.Col,
		UV:  t.V2.UV,
		Nor: t.V2.Nor.Apply(mvp.Normal),
	}
	t3 = &primitive.Vertex{
		Pos: mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(t.V3.Pos),
		Col: t.V3.Col,
		UV:  t.V3.UV,
		Nor: t.V3.Nor.Apply(mvp.Normal),
	}

	// For perspective corrected interpolation, see below.
	recipw := [3]float32{1, 1, 1}
	if r.cfg.Perspect {
		recipw = [3]float32{-1 / t1.Pos.W, -1 / t2.Pos.W, -1 / t3.Pos.W}
	}

	t1.Pos = t1.Pos.Apply(mvp.Viewport).Pos()
	t2.Pos = t2.Pos.Apply(mvp.Viewport).Pos()
	t3.Pos = t3.Pos.Apply(mvp.Viewport).Pos()
	if r.cullBackFace(t1.Pos, t2.Pos, t3.Pos) {
		return
	}
	if r.cullViewFrustum(r.bufs[0], t1.Pos, t2.Pos, t3.Pos) {
		return
	}

	if r.inViewport(r.bufs[0], t1.Pos, t2.Pos, t3.Pos) {
		r.drawClipped(mvp, t1, t2, t3, recipw, t.MaterialId)
		return
	}

	w := float32(r.cfg.MSAA * r.bufs[0].Bounds().Dx())
	h := float32(r.cfg.MSAA * r.bufs[0].Bounds().Dy())
	tris := r.clipTriangle(t1, t2, t3, w, h, recipw)
	for _, tri := range tris {
		r.drawClipped(mvp, tri.V1, tri.V2, tri.V3, recipw, t.MaterialId)
	}
}

func (r *Renderer) drawClipped(
	mvp *shader.MVP,
	t1, t2, t3 *primitive.Vertex,
	recipw [3]float32,
	materialId uint64) {
	m1 := t1.Pos.Apply(mvp.ViewportInv).Apply(mvp.ProjInv).Apply(mvp.ViewInv)
	m2 := t2.Pos.Apply(mvp.ViewportInv).Apply(mvp.ProjInv).Apply(mvp.ViewInv)
	m3 := t3.Pos.Apply(mvp.ViewportInv).Apply(mvp.ProjInv).Apply(mvp.ViewInv)

	// Compute AABB make the AABB a little bigger that align with
	// pixels to contain the entire triangle
	aabb := primitive.NewAABB(t1.Pos.ToVec3(), t2.Pos.ToVec3(), t3.Pos.ToVec3())
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	fN := m2.Sub(m1).Cross(m3.Sub(m1)).Unit()

	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if !r.bufs[0].In(x, y) {
				continue
			}

			p := math.NewVec2(float32(x)+0.5, float32(y)+0.5)
			bc := math.Barycoord(p, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())

			// Is inside triangle?
			if bc[0] < -math.Epsilon || bc[1] < -math.Epsilon || bc[2] < -math.Epsilon {
				continue
			}

			// Z-test
			z := bc[0]*t1.Pos.Z + bc[1]*t2.Pos.Z + bc[2]*t3.Pos.Z
			if !r.bufs[0].DepthTest(x, y, z) {
				continue
			}

			// Perspective corrected interpolation. See:
			// Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
			// Department of Computer Science, University of North Carolina at Chapel Hill (2002).
			wc1, wc2, wc3 := recipw[0]*bc[0], recipw[1]*bc[1], recipw[2]*bc[2]
			norm := float32(1.0)
			if r.cfg.Perspect {
				norm = 1 / (wc1 + wc2 + wc3)
			}

			// UV interpolation
			uvX := (wc1*t1.UV.X + wc2*t2.UV.X + wc3*t3.UV.X) * norm
			uvY := (wc1*t1.UV.Y + wc2*t2.UV.Y + wc3*t3.UV.Y) * norm

			// Compute du dv
			var du, dv float32
			if materialId > 0 {
				p1 := math.NewVec2(p.X+1, p.Y)
				p2 := math.NewVec2(p.X, p.Y+1)
				bcx := math.Barycoord(p1, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())
				wc1x, wc2x, wc3x := recipw[0]*bcx[0], recipw[1]*bcx[1], recipw[2]*bcx[2]
				normx := 1 / (wc1x + wc2x + wc3x)

				bcy := math.Barycoord(p2, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())
				wc1y, wc2y, wc3y := recipw[0]*bcy[0], recipw[1]*bcy[1], recipw[2]*bcy[2]
				normy := 1 / (wc1y + wc2y + wc3y)

				uvdU := (wc1x*t1.UV.X + wc2x*t2.UV.X + wc3x*t3.UV.X) * normx
				uvdX := (wc1x*t1.UV.Y + wc2x*t2.UV.Y + wc3x*t3.UV.Y) * normx

				uvdV := (wc1y*t1.UV.X + wc2y*t2.UV.X + wc3y*t3.UV.X) * normy
				uvdY := (wc1y*t1.UV.Y + wc2y*t2.UV.Y + wc3y*t3.UV.Y) * normy
				du = (uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY)
				dv = (uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY)
			}

			// normal interpolation (normals are in model space, no need for perspective correction)
			n := (math.Vec4[float32]{
				X: (bc[0]*t1.Nor.X + bc[1]*t2.Nor.X + bc[2]*t3.Nor.X),
				Y: (bc[0]*t1.Nor.Y + bc[1]*t2.Nor.Y + bc[2]*t3.Nor.Y),
				Z: (bc[0]*t1.Nor.Z + bc[1]*t2.Nor.Z + bc[2]*t3.Nor.Z),
				W: 0,
			}).Unit()
			pos := math.Vec4[float32]{
				X: (bc[0]*m1.X + bc[1]*m1.X + bc[2]*m1.X),
				Y: (bc[0]*m2.Y + bc[1]*m2.Y + bc[2]*m2.Y),
				Z: (bc[0]*m3.Z + bc[1]*m3.Z + bc[2]*m3.Z),
				W: 1,
			}
			col := color.RGBA{
				R: uint8(math.Clamp((wc1*float32(t1.Col.R)+wc2*float32(t2.Col.R)+wc3*float32(t3.Col.R))*norm, 0, 0xff)),
				G: uint8(math.Clamp((wc1*float32(t1.Col.G)+wc2*float32(t2.Col.G)+wc3*float32(t3.Col.G))*norm, 0, 0xff)),
				B: uint8(math.Clamp((wc1*float32(t1.Col.B)+wc2*float32(t2.Col.B)+wc3*float32(t3.Col.B))*norm, 0, 0xff)),
				A: uint8(math.Clamp((wc1*float32(t1.Col.A)+wc2*float32(t2.Col.A)+wc3*float32(t3.Col.A))*norm, 0, 0xff)),
			}

			// update G-buffer
			r.bufs[0].Set(x, y, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X:          x,
					Y:          y,
					Depth:      z,
					U:          uvX,
					V:          uvY,
					Du:         du,
					Dv:         dv,
					Nor:        n,
					Col:        col,
					MaterialID: materialId,
					AttrFlat: map[primitive.AttrName]any{
						"Pos": pos,
						"fN":  fN,
					},
				},
			})
		}
	}
}
