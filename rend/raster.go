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
	"golang.org/x/image/draw"
)

// Rasterizer is a CPU rasterizer
type Rasterizer struct {
	width          int
	height         int
	msaa           int
	concurrentSize int32 // atomic

	s       *Scene
	lockBuf []sync.Mutex
	gBuf    []rendInfo
	rendBuf *image.RGBA

	frameBuf *image.RGBA
}

type rendInfo struct {
	ok     bool
	z      float64
	u, v   float64
	du, dv float64
	n, pos math.Vector
	mat    material.Material
}

// NewRasterizer creates a new rasterizer
func NewRasterizer(width, height, msaa int) *Rasterizer {
	r := &Rasterizer{
		width:          width * msaa,
		height:         height * msaa,
		msaa:           msaa,
		gBuf:           make([]rendInfo, width*height*msaa*msaa),
		rendBuf:        image.NewRGBA(image.Rect(0, 0, width*msaa, height*msaa)),
		frameBuf:       image.NewRGBA(image.Rect(0, 0, width, height)),
		lockBuf:        make([]sync.Mutex, width*height*msaa*msaa),
		concurrentSize: 64, // empirical, see benchmark
	}
	r.resetBufs()
	return r
}

func (r *Rasterizer) RenderScene() *image.RGBA {
	if r.s == nil {
		panic("no scene!")
	}

	return r.Render(r.s)
}

func (r *Rasterizer) SetSize(width, height, msaa int) {
	r.width = width
	r.height = height
	r.lockBuf = make([]sync.Mutex, width*height*msaa*msaa)
}

func (r *Rasterizer) SetScene(s *Scene) {
	r.s = s
}

func (r *Rasterizer) resetBufs() {
	wg := sync.WaitGroup{}
	wg.Add(3)
	go func() {
		size := r.width * r.height / (r.msaa * r.msaa)
		for i := 0; i < size; i++ {
			r.frameBuf.Pix[i] = 0
		}
		wg.Done()
	}()
	go func() {
		size := r.width * r.height
		for i := 0; i < size; i++ {
			r.rendBuf.Pix[i] = 0
		}
		wg.Done()
	}()
	go func() {
		size := len(r.gBuf)
		for i := 0; i < size; i++ {
			r.gBuf[i].ok = false
		}
		wg.Done()
	}()
	wg.Wait()
}

// Render renders a scene.
func (r *Rasterizer) Render(s *Scene) *image.RGBA {
	r.s = s
	r.resetBufs()

	nP := runtime.GOMAXPROCS(0)
	limiter := utils.NewConccurLimiter(nP)

	for m := 0; m < len(r.s.Meshes); m++ {
		mesh := r.s.Meshes[m]

		uniforms := map[string]math.Matrix{
			"matModel":  mesh.ModelMatrix(),
			"matView":   r.s.Camera.ViewMatrix(),
			"matProj":   r.s.Camera.ProjMatrix(),
			"matVP":     math.ViewportMatrix(float64(r.width), float64(r.height)),
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

	limiter = utils.NewConccurLimiter(nP)
	xstep := int(r.concurrentSize)
	ystep := int(r.concurrentSize)
	for i := 0; i < r.width; i += xstep {
		for j := 0; j < r.height; j += ystep {
			ii := i
			jj := j
			limiter.Execute(func() {
				for k := 0; k < xstep; k++ {
					for l := 0; l < ystep; l++ {
						idx := ii + k + r.width*(jj+l)
						if idx >= len(r.gBuf) {
							continue
						}
						info := r.gBuf[idx]
						if !info.ok {
							continue
						}

						lod := math.Log2(math.Sqrt(math.Max(info.du, info.dv)))
						col := info.mat.Texture().Query(info.u, 1-info.v, lod)
						col = info.mat.Shader(col, info.pos, info.n, r.s.Lights[0].Position(), r.s.Camera.Position())
						r.fragmentProcessing(ii+k, jj+l, info.z, &col)
					}
				}
			})
		}
	}
	limiter.Wait()

	draw.BiLinear.Scale(r.frameBuf, r.frameBuf.Bounds(), r.rendBuf, image.Rectangle{
		image.Point{0, 0},
		image.Point{r.width, r.height},
	}, draw.Over, nil)
	return r.frameBuf
}

func (r *Rasterizer) barycoord(x, y int, t1, t2, t3 math.Vector) (w1, w2, w3 float64) {
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

func (r *Rasterizer) draw(uniforms map[string]math.Matrix, tri *geometry.Triangle, mat material.Material) {
	matModel := uniforms["matModel"]
	m1 := tri.V1.Position.Apply(matModel)
	m2 := tri.V1.Position.Apply(matModel)
	m3 := tri.V1.Position.Apply(matModel)

	t1 := r.vertexShader(tri.V1, uniforms)
	t2 := r.vertexShader(tri.V2, uniforms)
	t3 := r.vertexShader(tri.V3, uniforms)

	if t2.Position.Sub(t1.Position).Cross(t3.Position.Sub(t1.Position)).Z < 0 {
		return
	}
	if !r.inViewport(t1.Position, t2.Position, t3.Position) {
		return
	}

	// Perspective corrected interpolation
	t1Z := 1.0
	t2Z := 1.0
	t3Z := 1.0
	if _, ok := r.s.Camera.(camera.PerspectiveCamera); ok {
		t1Z = 1 / t1.Position.Z
		t2Z = 1 / t2.Position.Z
		t3Z = 1 / t3.Position.Z

		t1.UV = t1.UV.Scale(t1Z, t1Z, 0, 1)
		t2.UV = t2.UV.Scale(t2Z, t2Z, 0, 1)
		t3.UV = t3.UV.Scale(t3Z, t3Z, 0, 1)
		t1.Normal = t1.Normal.Scale(t1Z, t1Z, t1Z, 1)
		t2.Normal = t2.Normal.Scale(t2Z, t2Z, t2Z, 1)
		t3.Normal = t3.Normal.Scale(t3Z, t3Z, t3Z, 1)
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
			w1, w2, w3 := r.barycoord(x, y, t1.Position, t2.Position, t3.Position)

			// Is inside triangle?
			if w1 < 0 || w2 < 0 || w3 < 0 {
				continue
			}

			if x < 0 || x >= r.width {
				continue
			}

			if y < 0 || y >= r.height {
				continue
			}

			// Early Z-test
			z := w1*t1.Position.Z + w2*t2.Position.Z + w3*t3.Position.Z

			idx := x + y*r.width
			r.lockBuf[idx].Lock()
			if r.gBuf[idx].ok && z <= r.gBuf[idx].z {
				r.lockBuf[idx].Unlock()
				continue
			}
			r.lockBuf[idx].Unlock()

			// Perspective corrected interpolation. See:
			// Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
			// Department of Computer Science, University of North Carolina at Chapel Hill (2002).
			Z := 1.0
			if _, ok := r.s.Camera.(camera.PerspectiveCamera); ok {
				Z = w1*t1Z + w2*t2Z + w3*t3Z
			}

			// UV interpolation
			uvX := (w1*t1.UV.X + w2*t2.UV.X + w3*t3.UV.X) / Z
			uvY := (w1*t1.UV.Y + w2*t2.UV.Y + w3*t3.UV.Y) / Z

			// Compute du dv
			w1x, w2x, w3x := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
			w1y, w2y, w3y := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
			uvdU := (w1x*t1.UV.X + w2x*t2.UV.X + w3x*t3.UV.X) / Z
			uvdX := (w1x*t1.UV.Y + w2x*t2.UV.Y + w3x*t3.UV.Y) / Z
			uvdV := (w1y*t1.UV.X + w2y*t2.UV.X + w3y*t3.UV.X) / Z
			uvdY := (w1y*t1.UV.Y + w2y*t2.UV.Y + w3y*t3.UV.Y) / Z
			du := (uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY)
			dv := (uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY)

			// normal interpolation
			n := math.Vector{
				X: (w1*t1.Normal.X + w2*t2.Normal.X + w3*t3.Normal.X) / Z,
				Y: (w1*t1.Normal.Y + w2*t2.Normal.Y + w3*t3.Normal.Y) / Z,
				Z: (w1*t1.Normal.Z + w2*t2.Normal.Z + w3*t3.Normal.Z) / Z,
				W: 0,
			}
			pos := math.Vector{
				X: (w1*m1.X + w2*m1.X + w3*m1.X) / Z,
				Y: (w1*m2.Y + w2*m2.Y + w3*m2.Y) / Z,
				Z: (w1*m3.Z + w2*m3.Z + w3*m3.Z) / Z,
				W: 1,
			}

			r.lockBuf[idx].Lock()
			r.gBuf[idx].ok = true
			r.gBuf[idx].z = z
			r.gBuf[idx].u = uvX
			r.gBuf[idx].v = uvY
			r.gBuf[idx].du = du
			r.gBuf[idx].dv = dv
			r.gBuf[idx].n = n
			r.gBuf[idx].pos = pos
			r.gBuf[idx].mat = mat
			r.lockBuf[idx].Unlock()
		}
	}
}

func (r *Rasterizer) vertexShader(v geometry.Vertex, uniforms map[string]math.Matrix) geometry.Vertex {
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

func (r *Rasterizer) inViewport(v1, v2, v3 math.Vector) bool {

	viewportAABB := geometry.NewAABB(
		math.NewVector(float64(r.width), float64(r.height), 1, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 0, -1, 1),
	)
	triangleAABB := geometry.NewAABB(v1, v2, v3)
	return viewportAABB.Intersect(triangleAABB)
}

func (r *Rasterizer) fragmentProcessing(x, y int, z float64, col *color.RGBA) {
	if x < 0 || y >= r.width {
		return
	}
	if x < 0 || y >= r.height {
		return
	}

	idx := x + y*r.width
	r.lockBuf[idx].Lock()
	r.rendBuf.Set(x, r.height-y, col)
	r.lockBuf[idx].Unlock()
}

// SetConcurrencySize sets the number of triangles that is processed in parallel
func (r *Rasterizer) SetConcurrencySize(new int32) (old int32) {
	old = atomic.SwapInt32(&r.concurrentSize, new)
	return
}