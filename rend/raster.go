// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

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
		concurrentSize: 64,
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
			"matModel": mesh.ModelMatrix(),
			"matView":  matView,
			"matProj":  matProj,
			"matVP":    matVP,
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
					r.draw(uniforms, t, mesh.ModelMatrix(), m)
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

	blockSize := int(r.concurrentSize)
	wsteps := w / blockSize
	hsteps := h / blockSize

	r.workerPool.Add(uint64(wsteps*hsteps) + 2)
	for i := 0; i < wsteps*blockSize; i += blockSize {
		for j := 0; j < hsteps*blockSize; j += blockSize {
			ii := i
			jj := j
			r.workerPool.Execute(func() {
				for k := 0; k < blockSize; k++ {
					for l := 0; l < blockSize; l++ {
						x := ii + k
						y := jj + l

						r.shade(x, y, uniforms)
					}
				}
			})
		}
	}
	r.workerPool.Execute(func() {
		for i := wsteps * blockSize; i < w; i++ {
			for j := 0; j < hsteps*blockSize; j++ {
				r.shade(i, j, uniforms)
			}
		}
	})
	r.workerPool.Execute(func() {
		for i := 0; i < wsteps*blockSize; i++ {
			for j := hsteps * blockSize; j < h; j++ {
				r.shade(i, j, uniforms)
			}
		}
		for i := wsteps * blockSize; i < w; i++ {
			for j := hsteps * blockSize; j < h; j++ {
				r.shade(i, j, uniforms)
			}
		}
	})

	r.workerPool.Wait()
}

func (r *Renderer) shade(x, y int, uniforms map[string]interface{}) {
	w := r.width * r.msaa
	idx := x + w*y
	if idx >= len(r.gBuf) {
		return
	}
	info := &r.gBuf[idx]
	if !info.ok {
		r.setFramebuf(x, y, r.background)
		return
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

	if info.mat.AmbientOcclusion() {
		// FIXME: naive and super slow SSAO implementation. Optimize
		// when denoiser is avaliable.
		total := 0.0
		for a := 0.0; a < math.Pi*2-1e-4; a += math.Pi / 4 {
			total += math.Pi/2 - r.maxElevationAngle(x, y, math.Cos(a), math.Sin(a))
		}
		total /= (math.Pi / 2) * 8
		total = math.Pow(total, 10000)

		col = color.RGBA{
			uint8(total * float64(col.R)),
			uint8(total * float64(col.G)),
			uint8(total * float64(col.B)), col.A}
	}

	r.setFramebuf(x, y, col)
}

func (r *Renderer) maxElevationAngle(x, y int, dirX, dirY float64) float64 {
	p := math.NewVector(float64(x), float64(y), 0, 1)
	dir := math.NewVector(dirX, dirY, 0, 0)
	maxangle := 0.0
	w := float64(r.width * r.msaa)
	h := float64(r.height * r.msaa)
	for t := 0.0; t < 100; t += 1 {
		cur := p.Add(dir.Scale(t, t, 1, 1))
		if cur.X >= w || cur.Y >= h || cur.X < 0 || cur.Y < 0 {
			return maxangle
		}

		distance := p.Sub(cur).Len()
		if distance < 1 {
			continue
		}
		shadeIdx := int(cur.X) + int(w)*int(cur.Y)
		traceIdx := int(p.X) + int(w)*int(p.Y)

		elevation := r.gBuf[shadeIdx].z - r.gBuf[traceIdx].z
		maxangle = math.Max(maxangle, math.Atan(elevation/distance))
	}
	return maxangle
}

func (r *Renderer) passAntialiasing() {
	if r.debug {
		done := utils.Timed("antialiasing")
		defer done()
	}

	r.passGammaCorrect()
	r.outBuf = utils.Resize(r.width, r.height, r.frameBuf)
}

func (r *Renderer) setFramebuf(x, y int, c color.RGBA) {
	w := r.width * r.msaa
	h := r.height * r.msaa
	idx := x + y*w

	r.lockBuf[idx].Lock()
	r.frameBuf.Set(x, h-y, c)
	r.lockBuf[idx].Unlock()
}

func (r *Renderer) draw(
	uniforms map[string]interface{},
	tri *primitive.Triangle,
	modelMatrix math.Matrix,
	mat material.Material) {
	m1 := tri.V1.Pos.Apply(modelMatrix)
	m2 := tri.V2.Pos.Apply(modelMatrix)
	m3 := tri.V3.Pos.Apply(modelMatrix)

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

	w := r.width * r.msaa
	h := r.height * r.msaa
	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if x < 0 || x >= w || y < 0 || y >= h {
				continue
			}

			w1, w2, w3 := r.barycoord(x, y, t1.Pos, t2.Pos, t3.Pos)

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
				w1x, w2x, w3x := r.barycoord(x+1, y, t1.Pos, t2.Pos, t3.Pos)
				w1y, w2y, w3y := r.barycoord(x+1, y, t1.Pos, t2.Pos, t3.Pos)
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

	batch := 128 // empirical
	length := len(r.frameBuf.Pix)
	batcheEnd := length / (4 * batch)
	r.workerPool.Add(uint64(batcheEnd) + 1)

	for i := 0; i < batcheEnd*(4*batch); i += 4 * batch {
		offset := i
		r.workerPool.Execute(func() {
			for j := 0; j < 4*batch; j += 4 {
				r.frameBuf.Pix[offset+j+0] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[offset+j+0])/0xff)*0xff + 0.5)
				r.frameBuf.Pix[offset+j+1] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[offset+j+1])/0xff)*0xff + 0.5)
				r.frameBuf.Pix[offset+j+2] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[offset+j+2])/0xff)*0xff + 0.5)
			}
		})
	}
	r.workerPool.Execute(func() {
		for i := batcheEnd * (4 * batch); i < length; i += 4 {
			r.frameBuf.Pix[i+0] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[i+0])/0xff)*0xff + 0.5)
			r.frameBuf.Pix[i+1] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[i+1])/0xff)*0xff + 0.5)
			r.frameBuf.Pix[i+2] = uint8(color.FromLinear2sRGB(float64(r.frameBuf.Pix[i+2])/0xff)*0xff + 0.5)
		}
	})
	r.workerPool.Wait()
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

func (r *Renderer) barycoord(x, y int, t1, t2, t3 math.Vector) (w1, w2, w3 float64) {
	if t1.X == t2.X && t2.X == t3.X { // not a triangle
		return -1, -1, -1
	}
	if t1.Y == t2.Y && t2.Y == t3.Y { // not a triangle
		return -1, -1, -1
	}

	ap := math.Vector{X: float64(x) - t1.X, Y: float64(y) - t1.Y, Z: 0, W: 0}
	ab := math.Vector{X: t2.X - t1.X, Y: t2.Y - t1.Y, Z: 0, W: 0}
	ac := math.Vector{X: t3.X - t1.X, Y: t3.Y - t1.Y, Z: 0, W: 0}
	bc := math.Vector{X: t3.X - t2.X, Y: t3.Y - t2.Y, Z: 0, W: 0}
	bp := math.Vector{X: float64(x) - t2.X, Y: float64(y) - t2.Y, Z: 0, W: 0}
	Sabc := ab.Cross(ac).Z
	Sabp := ab.Cross(ap).Z
	Sapc := ap.Cross(ac).Z
	Sbcp := bc.Cross(bp).Z
	w1, w2, w3 = Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc
	return
}
