// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package rend

import (
	"image"
	"image/color"
	"runtime"
	"sync"
	"sync/atomic"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/utils"
)

// Renderer is a hybrid software renderer that implements
// rasterization and ray tracing.
type Renderer struct {
	// rendering options
	width        int
	height       int
	msaa         int
	useShadowMap bool
	debug        bool
	scene        *Scene
	background   color.RGBA

	// scheduling, use for hard interruption.
	running uint32 // atomic
	stop    uint32 // atomic

	// rendering caches
	concurrentSize int32
	lockBuf        []sync.Mutex
	gBuf           []gInfo
	frameBuf       *image.RGBA
	shadowTexture  []float64
	lightCamera    camera.OrthographicCamera
	outBuf         *image.RGBA
}

// NewRenderer creates a new renderer.
//
// The renderer implements a rasterization rendering pipeline.
func NewRenderer(opts ...Option) *Renderer {
	r := &Renderer{ // default settings
		width:        800,
		height:       500,
		msaa:         1,
		useShadowMap: false,
		debug:        true,
		scene:        nil,
	}
	for _, opt := range opts {
		opt(r)
	}

	// calibrate rendering size
	r.width *= r.msaa
	r.height *= r.msaa

	// initialize rendering caches
	r.concurrentSize = 64
	r.lockBuf = make([]sync.Mutex, r.width*r.height)
	r.gBuf = make([]gInfo, r.width*r.height)
	r.frameBuf = image.NewRGBA(image.Rect(0, 0, r.width, r.height))
	r.resetBufs()

	return r
}

func (r *Renderer) UpdateOptions(opts ...Option) {
	r.wait()

	for _, opt := range opts {
		opt(r)
	}

	// calibrate rendering size
	r.width *= r.msaa
	r.height *= r.msaa
	r.lockBuf = make([]sync.Mutex, r.width*r.height)
	r.gBuf = make([]gInfo, r.width*r.height)
	r.frameBuf = image.NewRGBA(image.Rect(0, 0, r.width, r.height))
	r.resetBufs()
}

// wait waits the current rendering terminates
func (r *Renderer) wait() {
	atomic.StoreUint32(&r.stop, 1)
	for atomic.LoadUint32(&r.running) == 0 {
		runtime.Gosched()
	}
	atomic.StoreUint32(&r.stop, 1)
}

func (r *Renderer) shouldStop() bool {
	return atomic.LoadUint32(&r.stop) == 1
}

func (r *Renderer) GetScene() *Scene {
	return r.scene
}

// Render renders a scene.
func (r *Renderer) Render() *image.RGBA {
	r.resetBufs()
	var (
		done       func()
		viewCamera camera.Interface
	)
	if r.shouldStop() {
		return r.outBuf
	}

	// shadow pass
	// TODO: compute optimal shadow map size
	if r.useShadowMap {
		r.lightCamera = camera.NewOrthographicCamera(
			r.scene.Lights[0].Position(),
			r.scene.Meshes[1].Center(),
			math.NewVector(0, 1, 0, 0),
			-0.3, 0.3, -0.2, 0.3, 0, -10,
		)
		viewCamera = r.scene.Camera
		r.scene.Camera = r.lightCamera
		if r.debug {
			done = utils.Timed("forward pass (shadow)....")
		}
		r.forwardPass()
		if r.debug {
			done()
		}

		r.shadowTexture = make([]float64, len(r.gBuf))
		for i, info := range r.gBuf {
			r.shadowTexture[i] = info.z
		}

		img := image.NewRGBA(image.Rect(0, 0, r.width, r.height))
		for i := 0; i < r.width; i++ {
			for j := 0; j < r.height; j++ {
				img.Set(i, j, color.RGBA{
					uint8(r.shadowTexture[i+(r.height-j-1)*r.width] * 255),
					uint8(r.shadowTexture[i+(r.height-j-1)*r.width] * 255),
					uint8(r.shadowTexture[i+(r.height-j-1)*r.width] * 255),
					255,
				})
			}
		}
		utils.Save(img, "shadow.png")
		r.scene.Camera = viewCamera
		r.resetBufs()
	}
	if r.shouldStop() {
		return r.outBuf
	}

	if r.debug {
		done = utils.Timed("forward pass (world)....")
	}
	r.forwardPass()
	if r.debug {
		done()
	}
	if r.shouldStop() {
		return r.outBuf
	}

	if r.debug {
		done = utils.Timed("deferred pass (shading)...")
	}
	r.deferredPass()
	if r.debug {
		done()
	}
	if r.shouldStop() {
		return r.outBuf
	}

	if r.debug {
		done = utils.Timed("antialiasing....")
	}
	r.antialiasing()
	if r.debug {
		done()
	}
	return r.outBuf
}

// gInfo is the geometry information collected in a forward pass.
type gInfo struct {
	ok     bool
	z      float64
	u, v   float64
	du, dv float64
	n, pos math.Vector
	mat    material.Material
}

func (r *Renderer) resetBufs() {
	for i := range r.frameBuf.Pix {
		r.frameBuf.Pix[i] = 0
	}
	for i := range r.gBuf {
		r.gBuf[i] = gInfo{z: -1}
	}
}

func (r *Renderer) forwardPass() {
	nP := runtime.GOMAXPROCS(0)
	limiter := utils.NewLimiter(nP)
	matView := r.scene.Camera.ViewMatrix()
	matProj := r.scene.Camera.ProjMatrix()
	matVP := math.ViewportMatrix(float64(r.width), float64(r.height))
	for m := range r.scene.Meshes {
		mesh := r.scene.Meshes[m]
		uniforms := map[string]math.Matrix{
			"matModel":  mesh.ModelMatrix(),
			"matView":   matView,
			"matProj":   matProj,
			"matVP":     matVP,
			"matNormal": mesh.NormalMatrix(),
		}

		length := len(mesh.Faces)
		for i := 0; i < length; i += int(r.concurrentSize) {
			ii := i
			limiter.Execute(func() {
				for k := int32(0); k < r.concurrentSize; k++ {
					if ii+int(k) >= length {
						return
					}

					r.draw(uniforms, mesh.Faces[ii+int(k)], mesh.Material)
				}
			})
		}
	}
	limiter.Wait()
}

func (r *Renderer) deferredPass() {
	nP := runtime.GOMAXPROCS(0)
	limiter := utils.NewLimiter(nP)
	xstep := int(r.concurrentSize)
	ystep := int(r.concurrentSize)

	matView := r.scene.Camera.ViewMatrix()
	matProj := r.scene.Camera.ProjMatrix()
	matVP := math.ViewportMatrix(float64(r.width), float64(r.height))
	for i := 0; i < r.width; i += xstep {
		for j := 0; j < r.height; j += ystep {
			ii := i
			jj := j
			limiter.Execute(func() {
				for k := 0; k < xstep; k++ {
					for l := 0; l < ystep; l++ {
						x := ii + k
						y := jj + l

						idx := x + r.width*y
						if idx >= len(r.gBuf) {
							continue
						}
						info := r.gBuf[idx]
						if !info.ok {
							r.setFramebuf(x, y, r.background)
							continue
						}

						lod := 0.0
						if info.mat.Texture().UseMipmap {
							lod = math.Log2(float64(info.mat.Texture().Size) * math.Sqrt(math.Max(info.du, info.dv)))
						}
						col := info.mat.Texture().Query(info.u, info.v, lod)
						col = info.mat.Shader(col, info.pos, info.n, r.scene.Camera.Position(), r.scene.Lights)

						if r.useShadowMap {
							// transform scrren coordinate to light viewport
							screenCoord := math.NewVector(float64(x), float64(y), info.z, 1).
								Apply(matVP.Inv()).
								Apply(matProj.Inv()).
								Apply(matView.Inv()).
								Apply(r.lightCamera.ViewMatrix()).
								Apply(r.lightCamera.ProjMatrix()).
								Apply(matVP)
							screenCoord = screenCoord.Scale(
								1/screenCoord.W,
								1/screenCoord.W,
								1/screenCoord.W,
								1/screenCoord.W,
							)

							// now the screend coordinates is transformed to
							// the light perspective, find out the depth we
							// need to query:
							lightX, lightY := int(screenCoord.X), int(screenCoord.Y)
							shadowIdx := lightX + lightY*r.width

							if shadowIdx > 0 && shadowIdx < len(r.shadowTexture) {
								shadowZ := r.shadowTexture[shadowIdx]

								// bilinear depth value query
								shadowIdx2 := lightX + 1 + lightY*r.width
								shadowIdx3 := lightX + (lightY+1)*r.width
								shadowIdx4 := lightX + 1 + (lightY+1)*r.width
								if (shadowIdx2 > 0 && shadowIdx2 < len(r.shadowTexture)) &&
									(shadowIdx3 > 0 && shadowIdx3 < len(r.shadowTexture)) &&
									(shadowIdx4 > 0 && shadowIdx4 < len(r.shadowTexture)) {

									shadowZ1 := shadowZ
									shadowZ2 := r.shadowTexture[shadowIdx2]
									shadowZ3 := r.shadowTexture[shadowIdx3]
									shadowZ4 := r.shadowTexture[shadowIdx4]
									tx := screenCoord.X - float64(lightX)
									shadowZa1 := math.Lerp(shadowZ1, shadowZ2, tx)
									shadowZa2 := math.Lerp(shadowZ3, shadowZ4, tx)
									ty := screenCoord.Y - float64(lightY)
									shadowZ = math.Lerp(shadowZa1, shadowZa2, ty)

									if screenCoord.Z < shadowZ-0.004 {
										col = color.RGBA{col.R / 2, col.G / 2, col.B / 2, 255}
									}
								}
							}
						}

						r.setFramebuf(x, y, col)
					}
				}
			})
		}
	}
	limiter.Wait()
}

func (r *Renderer) antialiasing() {
	r.outBuf = utils.Resize(r.width/r.msaa, r.height/r.msaa, r.frameBuf)
}

func (r *Renderer) setFramebuf(x, y int, c color.RGBA) {
	idx := x + y*r.width

	r.lockBuf[idx].Lock()
	r.frameBuf.Set(x, r.height-y, c)
	r.lockBuf[idx].Unlock()
}

func (r *Renderer) draw(uniforms map[string]math.Matrix, tri *geometry.Triangle, mat material.Material) {
	matModel := uniforms["matModel"]
	m1 := tri.V1.Position.Apply(matModel)
	m2 := tri.V1.Position.Apply(matModel)
	m3 := tri.V1.Position.Apply(matModel)

	t1 := r.vertexShader(tri.V1, uniforms)
	t2 := r.vertexShader(tri.V2, uniforms)
	t3 := r.vertexShader(tri.V3, uniforms)

	// Backface culling
	if t2.Position.Sub(t1.Position).Cross(t3.Position.Sub(t1.Position)).Z < 0 {
		return
	}

	// Viewfrustum culling
	if !r.inViewport(t1.Position, t2.Position, t3.Position) {
		return
	}

	// Perspective corrected interpolation
	t1Z := 1.0
	t2Z := 1.0
	t3Z := 1.0
	if _, ok := r.scene.Camera.(camera.PerspectiveCamera); ok {
		t1Z = 1 / t1.Position.Z
		t2Z = 1 / t2.Position.Z
		t3Z = 1 / t3.Position.Z

		t1.UV = t1.UV.Scale(t1Z, t1Z, 0, 1)
		t2.UV = t2.UV.Scale(t2Z, t2Z, 0, 1)
		t3.UV = t3.UV.Scale(t3Z, t3Z, 0, 1)
	}

	// Compute AABB make the AABB a little bigger that align with pixels
	// to contain the entire triangle
	aabb := geometry.NewAABB(t1.Position, t2.Position, t3.Position)
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if x < 0 || x >= r.width || y < 0 || y >= r.height {
				continue
			}

			w1, w2, w3 := r.barycoord(x, y, t1.Position, t2.Position, t3.Position)

			// Is inside triangle?
			if w1 < 0 || w2 < 0 || w3 < 0 {
				continue
			}

			// Z-test
			z := w1*t1.Position.Z + w2*t2.Position.Z + w3*t3.Position.Z
			if !r.passDepthTest(x, y, z) {
				continue
			}

			// Perspective corrected interpolation. See:
			// Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
			// Department of Computer Science, University of North Carolina at Chapel Hill (2002).
			Z := 1.0
			if _, ok := r.scene.Camera.(camera.PerspectiveCamera); ok {
				Z = w1*t1Z + w2*t2Z + w3*t3Z
			}

			// UV interpolation
			uvX := (w1*t1.UV.X + w2*t2.UV.X + w3*t3.UV.X) / Z
			uvY := (w1*t1.UV.Y + w2*t2.UV.Y + w3*t3.UV.Y) / Z

			// Compute du dv
			var du, dv float64
			if mat.Texture().UseMipmap {
				w1x, w2x, w3x := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
				w1y, w2y, w3y := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
				uvdU := (w1x*t1.UV.X + w2x*t2.UV.X + w3x*t3.UV.X) / Z
				uvdX := (w1x*t1.UV.Y + w2x*t2.UV.Y + w3x*t3.UV.Y) / Z
				uvdV := (w1y*t1.UV.X + w2y*t2.UV.X + w3y*t3.UV.X) / Z
				uvdY := (w1y*t1.UV.Y + w2y*t2.UV.Y + w3y*t3.UV.Y) / Z
				du = (uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY)
				dv = (uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY)
			}

			// normal interpolation
			n := (math.Vector{
				X: (w1*t1.Normal.X + w2*t2.Normal.X + w3*t3.Normal.X),
				Y: (w1*t1.Normal.Y + w2*t2.Normal.Y + w3*t3.Normal.Y),
				Z: (w1*t1.Normal.Z + w2*t2.Normal.Z + w3*t3.Normal.Z),
				W: 0,
			}).Unit()
			pos := math.Vector{
				X: (w1*m1.X + w2*m1.X + w3*m1.X),
				Y: (w1*m2.Y + w2*m2.Y + w3*m2.Y),
				Z: (w1*m3.Z + w2*m3.Z + w3*m3.Z),
				W: 1,
			}

			// update G-buffer
			idx := x + y*r.width
			r.lockBuf[idx].Lock()
			r.gBuf[idx].ok = true
			r.gBuf[idx].z = z
			r.gBuf[idx].u = uvX
			r.gBuf[idx].v = 1 - uvY
			r.gBuf[idx].du = du
			r.gBuf[idx].dv = dv
			r.gBuf[idx].n = n
			r.gBuf[idx].pos = pos
			r.gBuf[idx].mat = mat
			r.lockBuf[idx].Unlock()
		}
	}
}

func (r *Renderer) passDepthTest(x, y int, z float64) bool {
	idx := x + y*r.width

	r.lockBuf[idx].Lock()
	defer r.lockBuf[idx].Unlock()

	return !(r.gBuf[idx].ok && z <= r.gBuf[idx].z)
}

func (r *Renderer) vertexShader(v geometry.Vertex, uniforms map[string]math.Matrix) geometry.Vertex {
	matModel := uniforms["matModel"]
	matView := uniforms["matView"]
	matProj := uniforms["matProj"]
	matVP := uniforms["matVP"]
	matNormal := uniforms["matNormal"]

	pos := v.Position.Apply(matModel).Apply(matView).Apply(matProj).Apply(matVP)
	return geometry.Vertex{
		Position: pos.Scale(1/pos.W, 1/pos.W, 1/pos.W, 1/pos.W),
		Color:    v.Color,
		UV:       v.UV,
		Normal:   v.Normal.Apply(matNormal),
	}
}

func (r *Renderer) inViewport(v1, v2, v3 math.Vector) bool {
	viewportAABB := geometry.NewAABB(
		math.NewVector(float64(r.width), float64(r.height), 1, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 0, -1, 1),
	)
	triangleAABB := geometry.NewAABB(v1, v2, v3)
	return viewportAABB.Intersect(triangleAABB)
}

func (r *Renderer) barycoord(x, y int, t1, t2, t3 math.Vector) (w1, w2, w3 float64) {
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
