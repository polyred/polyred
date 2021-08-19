// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"fmt"
	"image"
	"runtime"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/object"
	"poly.red/scene"
	"poly.red/texture/buffer"
	"poly.red/texture/imageutil"

	"poly.red/internal/profiling"
	"poly.red/internal/sched"
)

// 1 second / 60fps = 16.6666 ms/frame
// 1 second / 30fps = 33.3333 ms/frame

// Renderer is a hybrid software renderer that implements
// rasterization and ray tracing.
type Renderer struct {
	// rendering options
	width        int
	height       int
	msaa         int
	correctGamma bool
	useShadowMap bool
	debug        bool
	scene        *scene.Scene
	background   color.RGBA
	blendFunc    BlendFunc

	// scheduling, use for hard interruption.
	running uint32 // atomic
	stop    uint32 // atomic

	// rendering caches
	lightSources   []light.Source
	lightEnv       []light.Environment
	batchSize      int32
	workers        int
	sched          *sched.Pool
	buf            *buffer.Buffer
	renderCamera   camera.Interface
	renderPerspect bool
	shadowBufs     []shadowInfo
	outBuf         *image.RGBA
}

// NewRenderer creates a new renderer.
//
// The returned renderer implements a rasterization rendering pipeline.
func NewRenderer(opts ...Opt) *Renderer {
	r := &Renderer{ // default settings
		buf:          nil,
		width:        800,
		height:       500,
		msaa:         1,
		useShadowMap: false,
		debug:        false,
		scene:        nil,
		workers:      runtime.NumCPU(),
		batchSize:    32, // heuristic
		lightSources: []light.Source{},
		lightEnv:     []light.Environment{},
	}
	for _, opt := range opts {
		opt(r)
	}

	r.buf = buffer.NewBuffer(image.Rect(0, 0, r.width*r.msaa, r.height*r.msaa))
	r.sched = sched.New(sched.Workers(r.workers))
	runtime.SetFinalizer(r, func(r *Renderer) {
		r.sched.Release()
	})

	if r.scene != nil {
		r.scene.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
			if o.Type() != object.TypeLight {
				return true
			}

			switch l := o.(type) {
			case light.Source:
				r.lightSources = append(r.lightSources, l)
			case light.Environment:
				r.lightEnv = append(r.lightEnv, l)
			}
			return true
		})
	}
	// initialize shadow maps
	if r.scene != nil && r.useShadowMap {
		r.initShadowMaps()
	}

	r.buf.ClearFragments()
	r.buf.ClearFrameBuf()
	return r
}

// Render renders a scene.
func (r *Renderer) Render() *image.RGBA {
	if r.debug {
		runtime.GOMAXPROCS(r.workers)
		fmt.Printf("rendering under GOMAXPROCS=%v\n", r.workers)
		total := profiling.Timed("entire rendering")
		defer total()
	}

	// record running
	r.startRunning()
	defer r.stopRunning()

	// reset buffers
	r.buf.ClearFragments()
	if r.shouldStop() {
		return r.outBuf
	}

	// decide if need shadow passes
	if r.useShadowMap {
		for i := 0; i < len(r.shadowBufs); i++ {
			r.passShadows(i)
			if r.shouldStop() {
				return r.outBuf
			}
		}
		r.buf.ClearFragments()
	}

	r.passForward()
	if r.shouldStop() {
		return r.outBuf
	}

	r.buf.ClearFrameBuf()
	r.passDeferred()
	if r.shouldStop() {
		return r.outBuf
	}

	r.passAntialiasing()
	return r.outBuf
}

// FrameBuf is a best effort way to return the current frame buffer.
func (r *Renderer) FrameBuf() *image.RGBA {
	if r.outBuf != nil {
		return r.outBuf
	}
	return r.buf.Image()
}

func (r *Renderer) passForward() {
	if r.debug {
		done := profiling.Timed("forward pass (world)")
		defer done()
	}

	matView := r.renderCamera.ViewMatrix()
	matProj := r.renderCamera.ProjMatrix()
	matVP := math.ViewportMatrix(float64(r.buf.Bounds().Dx()), float64(r.buf.Bounds().Dy()))

	r.scene.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
		if o.Type() != object.TypeMesh {
			return true
		}

		mesh := o.(mesh.Mesh)
		r.sched.Add(mesh.NumTriangles())
		return true
	})

	r.scene.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
		if o.Type() != object.TypeMesh {
			return true
		}

		mesh := o.(mesh.Mesh)
		uniforms := map[string]interface{}{
			"matModel":   mesh.ModelMatrix(),
			"matView":    matView,
			"matViewInv": matView.Inv(),
			"matProj":    matProj,
			"matProjInv": matProj.Inv(),
			"matVP":      matVP,
			"matVPInv":   matVP.Inv(),
			// NormalMatrix can be ((Tcamera * Tmodel)^(-1))^T or ((Tmodel)^(-1))^T
			// depending on which transformation space. Here we use the 2nd form,
			// i.e. model space normal matrix to save some computation of camera
			// transforamtion in the shading process.
			// The reason we need normal matrix is that normals are transformed
			// incorrectly using MVP matrices. However, a normal matrix helps us
			// to fix the problem.
			"matNormal": mesh.ModelMatrix().Inv().T(),
		}

		mesh.Faces(func(f primitive.Face, m material.Material) bool {
			f.Triangles(func(t *primitive.Triangle) bool {
				r.sched.Run(func() {
					if t.IsValid() {
						r.draw(uniforms, t, m)
					}
				})
				return true
			})
			return true
		})
		return true
	})
	r.sched.Wait()
}

func (r *Renderer) passDeferred() {
	if r.debug {
		done := profiling.Timed("deferred pass (shading)")
		defer done()
	}
	matView := r.renderCamera.ViewMatrix()
	matViewInv := matView.Inv()
	matProj := r.renderCamera.ProjMatrix()
	matProjInv := matProj.Inv()
	matVP := math.ViewportMatrix(float64(r.buf.Bounds().Dx()), float64(r.buf.Bounds().Dy()))
	matVPInv := matVP.Inv()
	matScreenToWorld := matViewInv.MulM(matProjInv).MulM(matVPInv)
	uniforms := map[string]interface{}{
		"matView":          matView,
		"matViewInv":       matViewInv,
		"matProj":          matProj,
		"matProjInv":       matProjInv,
		"matVP":            matVP,
		"matScreenToWorld": matScreenToWorld,
	}

	ao := ambientOcclusionPass{buf: r.buf}
	r.DrawFragments(r.buf, func(frag primitive.Fragment) color.RGBA {
		frag.Col = r.shade(r.buf.UnsafeAt(frag.X, frag.Y), uniforms)
		return ao.Shade(frag.X, frag.Y, frag.Col)
	})
}

func (r *Renderer) shade(info buffer.Fragment, uniforms map[string]interface{}) color.RGBA {
	if !info.Ok {
		return r.background
	}

	col := info.Col
	mat, ok := info.AttrFlat["Mat"].(material.Material)
	if !ok {
		mat = nil
	}

	pos := info.AttrFlat["Pos"].(math.Vec4)
	fN := info.AttrFlat["fN"].(math.Vec4)
	if mat != nil {
		lod := 0.0
		if mat.Texture().UseMipmap() {
			siz := float64(mat.Texture().Size()) * math.Sqrt(math.Max(info.Du, info.Dv))
			if siz < 1 {
				siz = 1
			}
			lod = math.Log2(siz)
		}

		col = mat.Texture().Query(lod, info.UV.X, 1-info.UV.Y)
		col = mat.FragmentShader(
			col, pos, info.Nor, fN,
			r.renderCamera.Position().ToVec4(1), r.lightSources, r.lightEnv)
	}

	if r.useShadowMap && mat != nil && mat.ReceiveShadow() {
		visibles := 0.0
		ns := len(r.shadowBufs)
		for i := 0; i < ns; i++ {
			visible := r.shadingVisibility(info.X, info.Y, i, info, uniforms)
			if visible {
				visibles++
			}
		}
		w := math.Pow(0.5, visibles)
		r := uint8(float64(col.R) * w)
		g := uint8(float64(col.G) * w)
		b := uint8(float64(col.B) * w)
		col = color.RGBA{r, g, b, col.A}
	}
	return col
}

func (r *Renderer) passAntialiasing() {
	if r.debug {
		done := profiling.Timed("antialiasing")
		defer done()
	}

	r.passGammaCorrect()
	r.outBuf = imageutil.Resize(r.width, r.height, r.buf.Image())
}

func (r *Renderer) draw(
	uniforms map[string]interface{},
	tri *primitive.Triangle,
	mat material.Material) {
	var t1, t2, t3 primitive.Vertex
	if mat != nil {
		t1 = mat.VertexShader(tri.V1, uniforms)
		t2 = mat.VertexShader(tri.V2, uniforms)
		t3 = mat.VertexShader(tri.V3, uniforms)
	} else {
		t1 = defaultVertexShader(tri.V1, uniforms)
		t2 = defaultVertexShader(tri.V2, uniforms)
		t3 = defaultVertexShader(tri.V3, uniforms)
	}

	matVP := uniforms["matVP"].(math.Mat4)

	// For perspective corrected interpolation, see below.
	recipw := math.NewVec4(1, 1, 1, 0)
	if _, ok := r.renderCamera.(*camera.Perspective); ok {
		recipw = math.NewVec4(-1/t1.Pos.W, -1/t2.Pos.W, -1/t3.Pos.W, 0)
	}

	t1.Pos = t1.Pos.Apply(matVP).Pos()
	t2.Pos = t2.Pos.Apply(matVP).Pos()
	t3.Pos = t3.Pos.Apply(matVP).Pos()

	// Backface culling
	if t2.Pos.Sub(t1.Pos).Cross(t3.Pos.Sub(t1.Pos)).Z < 0 {
		return
	}

	// Viewfrustum culling
	if !r.inViewport(t1.Pos, t2.Pos, t3.Pos) {
		return
	}

	w := float64(r.buf.Bounds().Dx())
	h := float64(r.buf.Bounds().Dy())

	// t1 is outside the viewfrustum
	outside := func(v *math.Vec4, w, h float64) bool {
		if v.X < 0 || v.X > w || v.Y < 0 || v.Y > h || v.Z > 1 || v.Z < -1 {
			return true
		}
		return false
	}

	if outside(&t1.Pos, w, h) || outside(&t2.Pos, w, h) || outside(&t3.Pos, w, h) {
		tris := r.clipTriangle(&t1, &t2, &t3, w, h, recipw)
		for _, tri := range tris {
			r.drawClipped(&tri.V1, &tri.V2, &tri.V3, recipw, uniforms, mat)
		}
		return
	}

	r.drawClipped(&t1, &t2, &t3, recipw, uniforms, mat)
}

func (r *Renderer) drawClipped(
	t1, t2, t3 *primitive.Vertex,
	recipw math.Vec4,
	uniforms map[string]interface{},
	mat material.Material) {

	matViewInv := uniforms["matViewInv"].(math.Mat4)
	matProjInv := uniforms["matProjInv"].(math.Mat4)
	matVPInv := uniforms["matVPInv"].(math.Mat4)
	m1 := t1.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)
	m2 := t2.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)
	m3 := t3.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)

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
			if !r.buf.In(x, y) {
				continue
			}

			p := math.NewVec2(float64(x)+0.5, float64(y)+0.5)
			bc := math.Barycoord(p, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())

			// Is inside triangle?
			if bc[0] < -math.Epsilon || bc[1] < -math.Epsilon || bc[2] < -math.Epsilon {
				continue
			}

			// Z-test
			z := bc[0]*t1.Pos.Z + bc[1]*t2.Pos.Z + bc[2]*t3.Pos.Z
			if !r.buf.DepthTest(x, y, z) {
				continue
			}

			// Perspective corrected interpolation. See:
			// Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
			// Department of Computer Science, University of North Carolina at Chapel Hill (2002).
			wc1, wc2, wc3 := recipw.X*bc[0], recipw.Y*bc[1], recipw.Z*bc[2]
			norm := 1.0
			if _, ok := r.renderCamera.(*camera.Perspective); ok {
				norm = 1 / (wc1 + wc2 + wc3)
			}

			// UV interpolation
			uvX := (wc1*t1.UV.X + wc2*t2.UV.X + wc3*t3.UV.X) * norm
			uvY := (wc1*t1.UV.Y + wc2*t2.UV.Y + wc3*t3.UV.Y) * norm

			// Compute du dv
			var du, dv float64
			if mat != nil && mat.Texture().UseMipmap() {
				p1 := math.NewVec2(p.X+1, p.Y)
				p2 := math.NewVec2(p.X, p.Y+1)
				bcx := math.Barycoord(p1, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())
				wc1x, wc2x, wc3x := recipw.X*bcx[0], recipw.Y*bcx[1], recipw.Z*bcx[2]
				normx := 1 / (wc1x + wc2x + wc3x)

				bcy := math.Barycoord(p2, t1.Pos.ToVec2(), t2.Pos.ToVec2(), t3.Pos.ToVec2())
				wc1y, wc2y, wc3y := recipw.X*bcy[0], recipw.Y*bcy[1], recipw.Z*bcy[2]
				normy := 1 / (wc1y + wc2y + wc3y)

				uvdU := (wc1x*t1.UV.X + wc2x*t2.UV.X + wc3x*t3.UV.X) * normx
				uvdX := (wc1x*t1.UV.Y + wc2x*t2.UV.Y + wc3x*t3.UV.Y) * normx

				uvdV := (wc1y*t1.UV.X + wc2y*t2.UV.X + wc3y*t3.UV.X) * normy
				uvdY := (wc1y*t1.UV.Y + wc2y*t2.UV.Y + wc3y*t3.UV.Y) * normy
				du = (uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY)
				dv = (uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY)
			}

			// normal interpolation (normals are in model space, no need for perspective correction)
			n := (math.Vec4{
				X: (bc[0]*t1.Nor.X + bc[1]*t2.Nor.X + bc[2]*t3.Nor.X),
				Y: (bc[0]*t1.Nor.Y + bc[1]*t2.Nor.Y + bc[2]*t3.Nor.Y),
				Z: (bc[0]*t1.Nor.Z + bc[1]*t2.Nor.Z + bc[2]*t3.Nor.Z),
				W: 0,
			}).Unit()
			pos := math.Vec4{
				X: (bc[0]*m1.X + bc[1]*m1.X + bc[2]*m1.X),
				Y: (bc[0]*m2.Y + bc[1]*m2.Y + bc[2]*m2.Y),
				Z: (bc[0]*m3.Z + bc[1]*m3.Z + bc[2]*m3.Z),
				W: 1,
			}
			col := color.RGBA{
				R: uint8(math.Clamp((wc1*float64(t1.Col.R)+wc2*float64(t2.Col.R)+wc3*float64(t3.Col.R))*norm, 0, 0xff)),
				G: uint8(math.Clamp((wc1*float64(t1.Col.G)+wc2*float64(t2.Col.G)+wc3*float64(t3.Col.G))*norm, 0, 0xff)),
				B: uint8(math.Clamp((wc1*float64(t1.Col.B)+wc2*float64(t2.Col.B)+wc3*float64(t3.Col.B))*norm, 0, 0xff)),
				A: uint8(math.Clamp((wc1*float64(t1.Col.A)+wc2*float64(t2.Col.A)+wc3*float64(t3.Col.A))*norm, 0, 0xff)),
			}

			// update G-buffer
			r.buf.Set(x, y, buffer.Fragment{
				Ok: true,
				Fragment: primitive.Fragment{
					X:     x,
					Y:     y,
					Depth: z,
					UV:    math.NewVec2(uvX, uvY),
					Du:    du,
					Dv:    dv,
					Nor:   n,
					Col:   col,
					AttrFlat: map[string]interface{}{
						"Pos": pos,
						"Mat": mat,
						"fN":  fN,
					},
				},
			})
		}
	}
}

// passGammaCorrect does a gamma correction that converts color from linear to sRGB space.
func (r *Renderer) passGammaCorrect() {
	if !r.correctGamma {
		return
	}

	r.DrawFragments(r.buf, func(frag primitive.Fragment) color.RGBA {
		r := uint8(color.FromLinear2sRGB(float64(frag.Col.R)/0xff)*0xff + 0.5)
		g := uint8(color.FromLinear2sRGB(float64(frag.Col.G)/0xff)*0xff + 0.5)
		b := uint8(color.FromLinear2sRGB(float64(frag.Col.B)/0xff)*0xff + 0.5)
		return color.RGBA{r, g, b, frag.Col.A}
	})
}

func (r *Renderer) inViewport(v1, v2, v3 math.Vec4) bool {
	viewportAABB := primitive.NewAABB(
		math.NewVec3(float64(r.buf.Bounds().Dx()), float64(r.buf.Bounds().Dy()), 1),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 0, -1),
	)
	triangleAABB := primitive.NewAABB(v1.ToVec3(), v2.ToVec3(), v3.ToVec3())
	return viewportAABB.Intersect(triangleAABB)
}
