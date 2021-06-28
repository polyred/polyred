// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"fmt"
	"image"
	"runtime"
	"sync"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
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

	// scheduling, use for hard interruption.
	running uint32 // atomic
	stop    uint32 // atomic

	// rendering caches
	lightSources   []light.Source
	lightEnv       []light.Environment
	concurrentSize int32
	gomaxprocs     int
	workerPool     *utils.WorkerPool
	lockBuf        []sync.Mutex
	gBuf           []gInfo
	frameBuf       *image.RGBA
	renderCamera   camera.Interface
	shadowBufs     []shadowInfo
	outBuf         *image.RGBA
}

// NewRenderer creates a new renderer.
//
// The renderer implements a rasterization rendering pipeline.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{ // default settings
		width:          800,
		height:         500,
		msaa:           1,
		useShadowMap:   false,
		debug:          false,
		scene:          nil,
		gomaxprocs:     runtime.GOMAXPROCS(0),
		concurrentSize: 32,
		lightSources:   []light.Source{},
		lightEnv:       []light.Environment{},
	}
	for _, opt := range opts {
		opt(r)
	}

	w := r.width * r.msaa
	h := r.height * r.msaa

	// initialize rendering caches
	r.lockBuf = make([]sync.Mutex, w*h)
	r.gBuf = make([]gInfo, w*h)
	r.frameBuf = image.NewRGBA(image.Rect(0, 0, w, h))
	r.workerPool = utils.NewWorkerPool(uint64(r.gomaxprocs))

	if r.scene != nil {
		r.scene.IterObjects(func(o object.Object, modelMatrix math.Matrix) bool {
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

	r.resetGBuf()
	r.resetFrameBuf()
	return r
}

// Render renders a scene.
func (r *Renderer) Render() *image.RGBA {
	if r.debug {
		runtime.GOMAXPROCS(r.gomaxprocs)
		fmt.Printf("rendering under GOMAXPROCS=%v\n", r.gomaxprocs)
		total := utils.Timed("entire rendering")
		defer total()
	}

	// record running
	r.startRunning()
	defer r.stopRunning()

	// reset buffers
	r.resetGBuf()
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
		r.resetGBuf()
	}

	r.passForward()
	if r.shouldStop() {
		return r.outBuf
	}

	r.resetFrameBuf()
	r.passDeferred()
	if r.shouldStop() {
		return r.outBuf
	}

	r.passAntialiasing()
	return r.outBuf
}

// gInfo is the geometry information collected in a forward pass.
type gInfo struct {
	ok         bool
	z          float64
	u, v       float64
	du, dv     float64
	n, fN, pos math.Vector
	col        color.RGBA
	mat        material.Material
}

func (r *Renderer) passForward() {
	if r.debug {
		done := utils.Timed("forward pass (world)")
		defer done()
	}

	w := r.width * r.msaa
	h := r.height * r.msaa
	r.renderCamera = r.scene.GetCamera()
	matView := r.renderCamera.ViewMatrix()
	matProj := r.renderCamera.ProjMatrix()
	matVP := math.ViewportMatrix(float64(w), float64(h))

	r.scene.IterObjects(func(o object.Object, modelMatrix math.Matrix) bool {
		if o.Type() != object.TypeMesh {
			return true
		}

		mesh := o.(geometry.Mesh)
		r.workerPool.Add(mesh.NumTriangles())
		return true
	})

	r.scene.IterObjects(func(o object.Object, modelMatrix math.Matrix) bool {
		if o.Type() != object.TypeMesh {
			return true
		}

		mesh := o.(geometry.Mesh)
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
				r.workerPool.Execute(func() {
					if t.IsValid() {
						r.draw(uniforms, t, mesh.ModelMatrix(), m)
					}
				})
				return true
			})
			return true
		})
		return true
	})
	r.workerPool.Wait()
}

func (r *Renderer) passDeferred() {
	if r.debug {
		done := utils.Timed("deferred pass (shading)")
		defer done()
	}
	w := r.width * r.msaa
	h := r.height * r.msaa
	r.renderCamera = r.scene.GetCamera()
	matView := r.renderCamera.ViewMatrix()
	matViewInv := matView.Inv()
	matProj := r.renderCamera.ProjMatrix()
	matProjInv := matProj.Inv()
	matVP := math.ViewportMatrix(float64(w), float64(h))
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

	ao := ambientOcclusionPass{
		w:       r.width * r.msaa,
		h:       r.height * r.msaa,
		gbuffer: r.gBuf,
	}

	r.ScreenPass(r.frameBuf, func(x, y int, col color.RGBA) color.RGBA {
		col = r.shade(x, h-y-1, uniforms)
		return ao.Shade(x, h-y-1, col)
	})
}

func (r *Renderer) shade(x, y int, uniforms map[string]interface{}) color.RGBA {
	w := r.width * r.msaa
	idx := x + w*y
	info := &r.gBuf[idx]
	if !info.ok {
		return r.background
	}

	col := info.col
	if info.mat != nil {
		lod := 0.0
		if info.mat.Texture().UseMipmap() {
			siz := float64(info.mat.Texture().Size()) * math.Sqrt(math.Max(info.du, info.dv))
			if siz < 1 {
				siz = 1
			}
			lod = math.Log2(siz)
		}

		col = info.mat.Texture().Query(lod, info.u, 1-info.v)
		col = info.mat.FragmentShader(
			col, info.pos, info.n, info.fN,
			r.renderCamera.Position(), r.lightSources, r.lightEnv)
	}

	if r.useShadowMap && info.mat != nil && info.mat.ReceiveShadow() {
		visibles := 0.0
		ns := len(r.shadowBufs)
		for i := 0; i < ns; i++ {
			visible := r.shadingVisibility(x, y, i, info, uniforms)
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
		done := utils.Timed("antialiasing")
		defer done()
	}

	r.passGammaCorrect()
	r.outBuf = utils.Resize(r.width, r.height, r.frameBuf)
}

func (r *Renderer) draw(
	uniforms map[string]interface{},
	tri *primitive.Triangle,
	modelMatrix math.Matrix,
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

	// Backface culling
	if t2.Pos.Sub(t1.Pos).Cross(t3.Pos.Sub(t1.Pos)).Z < 0 {
		return
	}

	// Viewfrustum culling
	if !r.inViewport(t1.Pos, t2.Pos, t3.Pos) {
		return
	}

	w := r.width * r.msaa
	h := r.height * r.msaa

	// t1 is outside the viewfrustum
	outside := func(v *math.Vector, w, h float64) bool {
		if v.X < 0 || v.X > w || v.Y < 0 || v.Y > h || v.Z > 1 || v.Z < -1 {
			return true
		}
		return false
	}

	if outside(&t1.Pos, float64(w), float64(h)) || outside(&t2.Pos, float64(w), float64(h)) || outside(&t3.Pos, float64(w), float64(h)) {
		tris := r.clipTriangle(&t1, &t2, &t3, float64(w), float64(h))
		for _, tri := range tris {
			r.drawClipped(&tri.V1, &tri.V2, &tri.V3, uniforms, mat)
		}
		return
	}

	r.drawClipped(&t1, &t2, &t3, uniforms, mat)
}
func (r *Renderer) drawClipped(
	t1, t2, t3 *primitive.Vertex,
	uniforms map[string]interface{},
	mat material.Material) {

	matViewInv := uniforms["matViewInv"].(math.Matrix)
	matProjInv := uniforms["matProjInv"].(math.Matrix)
	matVPInv := uniforms["matVPInv"].(math.Matrix)
	m1 := t1.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)
	m2 := t2.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)
	m3 := t3.Pos.Apply(matVPInv).Apply(matProjInv).Apply(matViewInv)

	w := r.width * r.msaa
	h := r.height * r.msaa

	// Perspective corrected interpolation
	t1Z := 1.0
	t2Z := 1.0
	t3Z := 1.0
	if _, ok := r.renderCamera.(*camera.Perspective); ok {
		t1Z = 1 / t1.Pos.Z
		t2Z = 1 / t2.Pos.Z
		t3Z = 1 / t3.Pos.Z

		t1.UV = t1.UV.Scale(t1Z, t1Z, 0, 1)
		t2.UV = t2.UV.Scale(t2Z, t2Z, 0, 1)
		t3.UV = t3.UV.Scale(t3Z, t3Z, 0, 1)
	}

	// Compute AABB make the AABB a little bigger that align with pixels
	// to contain the entire triangle
	aabb := primitive.NewAABB(t1.Pos, t2.Pos, t3.Pos)
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	fN := m2.Sub(m1).Cross(m3.Sub(m1)).Unit()

	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if x < 0 || x >= w || y < 0 || y >= h {
				continue
			}

			w1, w2, w3 := math.Barycoord(math.NewVector(float64(x), float64(y), 0, 1),
				t1.Pos, t2.Pos, t3.Pos)

			// Is inside triangle?
			if w1 < 0 || w2 < 0 || w3 < 0 {
				continue
			}

			// Z-test
			z := w1*t1.Pos.Z + w2*t2.Pos.Z + w3*t3.Pos.Z
			if !r.depthTest(x, y, z) {
				continue
			}

			// Perspective corrected interpolation. See:
			// Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
			// Department of Computer Science, University of North Carolina at Chapel Hill (2002).
			Z := 1.0
			if _, ok := r.renderCamera.(*camera.Perspective); ok {
				Z = w1*t1Z + w2*t2Z + w3*t3Z
			}

			// UV interpolation
			uvX := (w1*t1.UV.X + w2*t2.UV.X + w3*t3.UV.X) / Z
			uvY := (w1*t1.UV.Y + w2*t2.UV.Y + w3*t3.UV.Y) / Z

			// Compute du dv
			var du, dv float64
			if mat != nil && mat.Texture().UseMipmap() {
				w1x, w2x, w3x := math.Barycoord(math.NewVector(float64(x+1), float64(y), 0, 1), t1.Pos, t2.Pos, t3.Pos)
				w1y, w2y, w3y := math.Barycoord(math.NewVector(float64(x), float64(y+1), 0, 1), t1.Pos, t2.Pos, t3.Pos)
				uvdU := (w1x*t1.UV.X + w2x*t2.UV.X + w3x*t3.UV.X) / Z
				uvdX := (w1x*t1.UV.Y + w2x*t2.UV.Y + w3x*t3.UV.Y) / Z
				uvdV := (w1y*t1.UV.X + w2y*t2.UV.X + w3y*t3.UV.X) / Z
				uvdY := (w1y*t1.UV.Y + w2y*t2.UV.Y + w3y*t3.UV.Y) / Z
				du = (uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY)
				dv = (uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY)
			}

			// normal interpolation
			n := (math.Vector{
				X: (w1*t1.Nor.X + w2*t2.Nor.X + w3*t3.Nor.X),
				Y: (w1*t1.Nor.Y + w2*t2.Nor.Y + w3*t3.Nor.Y),
				Z: (w1*t1.Nor.Z + w2*t2.Nor.Z + w3*t3.Nor.Z),
				W: 0,
			}).Unit()
			pos := math.Vector{
				X: (w1*m1.X + w2*m1.X + w3*m1.X),
				Y: (w1*m2.Y + w2*m2.Y + w3*m2.Y),
				Z: (w1*m3.Z + w2*m3.Z + w3*m3.Z),
				W: 1,
			}
			col := color.RGBA{
				R: uint8(math.Clamp(w1*float64(t1.Col.R)+w2*float64(t2.Col.R)+w3*float64(t3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(w1*float64(t1.Col.G)+w2*float64(t2.Col.G)+w3*float64(t3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(w1*float64(t1.Col.B)+w2*float64(t2.Col.B)+w3*float64(t3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(w1*float64(t1.Col.A)+w2*float64(t2.Col.A)+w3*float64(t3.Col.A), 0, 0xff)),
			}

			// update G-buffer
			idx := x + y*w
			r.lockBuf[idx].Lock()
			r.gBuf[idx].ok = true
			r.gBuf[idx].z = z
			r.gBuf[idx].u = uvX
			r.gBuf[idx].v = uvY
			r.gBuf[idx].du = du
			r.gBuf[idx].dv = dv
			r.gBuf[idx].n = n
			r.gBuf[idx].fN = fN
			r.gBuf[idx].pos = pos
			r.gBuf[idx].col = col
			r.gBuf[idx].mat = mat
			r.lockBuf[idx].Unlock()
		}
	}
}

// passGammaCorrect does a gamma correction that converts color from linear to sRGB space.
func (r *Renderer) passGammaCorrect() {
	if !r.correctGamma {
		return
	}

	r.ScreenPass(r.frameBuf, func(x, y int, col color.RGBA) color.RGBA {
		r := uint8(color.FromLinear2sRGB(float64(col.R)/0xff)*0xff + 0.5)
		g := uint8(color.FromLinear2sRGB(float64(col.G)/0xff)*0xff + 0.5)
		b := uint8(color.FromLinear2sRGB(float64(col.B)/0xff)*0xff + 0.5)
		return color.RGBA{r, g, b, col.A}
	})
}

func (r *Renderer) depthTest(x, y int, z float64) bool {
	w := r.width * r.msaa
	idx := x + y*w

	r.lockBuf[idx].Lock()
	defer r.lockBuf[idx].Unlock()
	return !(r.gBuf[idx].ok && z <= r.gBuf[idx].z)
}

func (r *Renderer) inViewport(v1, v2, v3 math.Vector) bool {
	viewportAABB := primitive.NewAABB(
		math.NewVector(float64(r.width*r.msaa), float64(r.height*r.msaa), 1, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 0, -1, 1),
	)
	triangleAABB := primitive.NewAABB(v1, v2, v3)
	return viewportAABB.Intersect(triangleAABB)
}
