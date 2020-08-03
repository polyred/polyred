// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
)

// Rasterizer is a CPU rasterizer
type Rasterizer struct {
	c *PerspectiveCamera // camera
	s *Scene

	width          int
	height         int
	frameBuf       []color.RGBA
	depthBuf       []float64
	lockBuf        []sync.Mutex
	concurrentSize int32 // atomic

	viewMatrix     Matrix
	projMatrix     Matrix
	viewportMatrix Matrix
}

// NewRasterizer creates a new rasterizer
func NewRasterizer(width, height int) *Rasterizer {
	return &Rasterizer{
		width:          width,
		height:         height,
		frameBuf:       make([]color.RGBA, width*height),
		depthBuf:       make([]float64, width*height),
		lockBuf:        make([]sync.Mutex, width*height),
		concurrentSize: 128, // empirical, see benchmark
		viewMatrix:     NewMatrix(),
		projMatrix:     NewMatrix(),
		viewportMatrix: NewMatrix(),
	}
}

// SetCamera sets the rasterizer camera
func (r *Rasterizer) SetCamera(c *PerspectiveCamera) {
	r.c = c
}

// SetScene sets the rasterizer scene
func (r *Rasterizer) SetScene(s *Scene) {
	r.s = s
}

// SetConcurrencySize sets the number of triangles that is processed in parallel
func (r *Rasterizer) SetConcurrencySize(new int32) (old int32) {
	old = atomic.SwapInt32(&r.concurrentSize, new)
	return
}

// Render renders a scene graph
func (r *Rasterizer) Render() {
	r.initBufs()
	r.initTrans()
	limiter := NewConccurLimiter(runtime.GOMAXPROCS(0))
	for _, o := range r.s.Objects {
		o.modelMatrix.SetIdentity().MultiplyMatrices(&o.translateMatrix, &o.scaleMatrix)
		o.normalMatrix.SetIdentity().MultiplyMatrix(&o.modelMatrix).Inverse().Transpose()
		for i := 0; i < len(o.triangles); i += int(r.concurrentSize) {
			ii := i
			limiter.Execute(func() {
				for k := int32(0); k < r.concurrentSize; k++ {
					if ii+int(k) >= len(o.triangles) {
						return
					}
					r.draw(o.triangles[ii+int(k)], o.texture, &o.modelMatrix, &o.normalMatrix)
				}
			})
		}
	}
	limiter.Wait()
}

// Save stores the current frame buffer to a newly created file.
func (r *Rasterizer) Save(filename string) {
	err := r.flushFrameBuffer(filename)
	if err != nil {
		panic(fmt.Errorf("cannot save the render result, err: %v", err))
	}
}

// flushFrameBuffer writes the frame buffer to an image
func (r *Rasterizer) flushFrameBuffer(filename string) error {
	m := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{r.width, r.height}})
	for i := 0; i < r.width; i++ {
		for j := 0; j < r.height; j++ {
			m.Set(i, r.height-j, r.frameBuf[j*r.width+i])
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	err = png.Encode(f, m)
	if err != nil {
		return err
	}

	return nil
}

func (r *Rasterizer) draw(tri *Triangle, tex *Texture, modelMatrix, normalMatrix *Matrix) {
	v1 := r.VertexShader(tri.v1, modelMatrix)
	v2 := r.VertexShader(tri.v2, modelMatrix)
	v3 := r.VertexShader(tri.v3, modelMatrix)

	// backface culling
	f1 := (&Vector{v2.Position.X, v2.Position.Y, v2.Position.Z, 1}).Sub(&v1.Position)
	f2 := (&Vector{v3.Position.X, v3.Position.Y, v3.Position.Z, 1}).Sub(&v1.Position)
	fN := (&Vector{}).CrossVectors(f1, f2)
	if (&Vector{0, 0, -1, 0}).Dot(fN) >= 0 {
		return
	}

	box := NewAABB(v1, v2, v3)

	// view frustum culling
	xMax := math.Min(box.max.X, float64(r.width))
	xMin := math.Max(box.min.X, 0)
	yMax := math.Min(box.max.Y, float64(r.height))
	yMin := math.Max(box.min.Y, 0)
	if xMin > xMax && yMin > yMax {
		return
	}

	// compute normals and shading point in world space for fragment shading
	n1 := (&Vector{tri.v1.Normal.X, tri.v1.Normal.Y, tri.v1.Normal.Z, 0}).
		ApplyMatrix(normalMatrix)
	n2 := (&Vector{tri.v2.Normal.X, tri.v2.Normal.Y, tri.v2.Normal.Z, 0}).
		ApplyMatrix(normalMatrix)
	n3 := (&Vector{tri.v3.Normal.X, tri.v3.Normal.Y, tri.v3.Normal.Z, 0}).
		ApplyMatrix(normalMatrix)
	a := (&Vector{tri.v1.Position.X, tri.v1.Position.Y, tri.v1.Position.Z, 1}).
		ApplyMatrix(modelMatrix)
	b := (&Vector{tri.v2.Position.X, tri.v2.Position.Y, tri.v2.Position.Z, 1}).
		ApplyMatrix(modelMatrix)
	c := (&Vector{tri.v3.Position.X, tri.v3.Position.Y, tri.v3.Position.Z, 1}).
		ApplyMatrix(modelMatrix)

	for x := math.Floor(xMin); x < xMax; x++ {
		for y := math.Floor(yMin); y < yMax; y++ {
			// compute barycentric
			ap := (&Vector{x, y, 0, 1}).
				Sub(&Vector{v1.Position.X, v1.Position.Y, 0, 1})
			ab := (&Vector{v2.Position.X, v2.Position.Y, 0, 1}).
				Sub(&Vector{v1.Position.X, v1.Position.Y, 0, 1})
			ac := (&Vector{v3.Position.X, v3.Position.Y, 0, 1}).
				Sub(&Vector{v1.Position.X, v1.Position.Y, 0, 1})
			bc := (&Vector{v3.Position.X, v3.Position.Y, 0, 1}).
				Sub(&Vector{v2.Position.X, v2.Position.Y, 0, 1})
			bp := (&Vector{x, y, 0, 1}).
				Sub(&Vector{v2.Position.X, v2.Position.Y, 0, 1})
			out := &Vector{0, 0, 1, 0}
			Sabc := (&Vector{}).CrossVectors(ab, ac).Dot(out)
			Sabp := (&Vector{}).CrossVectors(ab, ap).Dot(out)
			Sapc := (&Vector{}).CrossVectors(ap, ac).Dot(out)
			Sbcp := (&Vector{}).CrossVectors(bc, bp).Dot(out)
			w1, w2, w3 := Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc

			// skip frags outside the triangle
			if w1 < 0 || w2 < 0 || w3 < 0 {
				continue
			}

			z := w1*v1.Position.Z + w2*v2.Position.Z + w3*v3.Position.Z
			idx := int(y*float64(r.width) + x)
			r.lockBuf[idx].Lock()
			if z < r.depthBuf[idx] {
				r.lockBuf[idx].Unlock()
				continue
			}
			r.lockBuf[idx].Unlock()

			// uv interpolation
			uv := &Vector{
				w1*tri.v1.UV.X + w2*tri.v2.UV.X + w3*tri.v3.UV.X,
				w1*tri.v1.UV.Y + w2*tri.v2.UV.Y + w3*tri.v3.UV.Y,
				0, 1,
			}
			p := &Vector{
				w1*a.X + w2*b.X + w3*c.X,
				w1*a.Y + w2*b.Y + w3*c.Y,
				w1*a.Z + w2*b.Z + w3*c.Z, 1,
			}
			n := &Vector{
				w1*n1.X + w2*n2.X + w3*n3.X,
				w1*n1.Y + w2*n2.Y + w3*n3.Y,
				w1*n1.Z + w2*n2.Z + w3*n3.Z, 0,
			}
			n.Normalize()
			c := r.FragmentShader(tex, uv, n, p)

			r.lockBuf[idx].Lock()
			r.depthBuf[idx] = z
			r.frameBuf[idx] = c
			r.lockBuf[idx].Unlock()
		}
	}
}

func (r *Rasterizer) initBufs() {
	size := r.width * r.height
	for i := 0; i < size; i++ {
		r.frameBuf[i] = color.RGBA{0, 0, 0, 0}
		r.depthBuf[i] = -math.MaxInt64
	}
}

func (r *Rasterizer) initTrans() {
	w := &Vector{r.c.LookAt.X, r.c.LookAt.Y, r.c.LookAt.Z, 1}
	w.Sub(&r.c.Position).Normalize()
	u := &Vector{}
	u.CrossVectors(&r.c.Up, w).MultiplyScalar(-1).Normalize()
	vv := &Vector{}
	vv.CrossVectors(u, w).Normalize()
	r.viewMatrix.Set(
		u.X, u.Y, u.Z, -r.c.Position.X*u.X-r.c.Position.Y*u.Y-r.c.Position.Z*u.Z,
		vv.X, vv.Y, vv.Z, -r.c.Position.X*vv.X-r.c.Position.Y*vv.Y-r.c.Position.Z*vv.Z,
		-w.X, -w.Y, -w.Z, r.c.Position.X*w.X+r.c.Position.Y*w.Y+r.c.Position.Z*w.Z,
		0, 0, 0, 1,
	)

	r.projMatrix.Set(
		-1/(r.c.Aspect*math.Tan(r.c.FOV*math.Pi/360)), 0, 0, 0,
		0, -1/(math.Tan(r.c.FOV*math.Pi/360)), 0, 0,
		0, 0, (r.c.Near+r.c.Far)/(r.c.Near-r.c.Far),
		2*(r.c.Near*r.c.Far)/(r.c.Near-r.c.Far),
		0, 0, 1, 0,
	)

	r.viewportMatrix.Set(
		float64(r.width)/2, 0, 0, float64(r.width)/2,
		0, float64(r.height)/2, 0, float64(r.height)/2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	)
}

// VertexShader ...
func (r *Rasterizer) VertexShader(v Vertex, modelMatrix *Matrix) Vertex {
	v.Position.ApplyMatrix(modelMatrix).ApplyMatrix(&r.viewMatrix).
		ApplyMatrix(&r.projMatrix).ApplyMatrix(&r.viewportMatrix)
	v.Position.X /= v.Position.W
	v.Position.Y /= v.Position.W
	v.Position.Z /= v.Position.W
	v.Position.W = 1
	return v
}

// FragmentShader ...
func (r *Rasterizer) FragmentShader(tex *Texture, uv, normal, p *Vector) color.RGBA {
	width := float64(tex.width)
	height := float64(tex.height)
	R, G, B, A := tex.data.At(
		int(math.Floor(uv.X*width)), int(height-math.Floor(uv.Y*height))).RGBA()
	I := color.RGBA{uint8(R >> 8), uint8(G >> 8), uint8(B >> 8), uint8(A >> 8)}
	L := (&Vector{r.s.Lights[0].Position.X, r.s.Lights[0].Position.Y,
		r.s.Lights[0].Position.Z, 1}).Sub(p).Normalize()
	V := (&Vector{r.c.Position.X, r.c.Position.Y, r.c.Position.Z, 1}).
		Sub(p).Normalize()
	H := (&Vector{L.X, L.Y, L.Z, 0}).
		Add(V).Normalize()
	La := clamp((&Vector{float64(I.R), float64(I.G), float64(I.B), 0}).
		MultiplyScalar(r.s.Lights[0].Kamb), 0, 255)
	Ld := clamp((&Vector{float64(I.R), float64(I.G), float64(I.B), 0}).
		MultiplyScalar(r.s.Lights[0].Kdiff).
		MultiplyScalar(normal.Dot(L)), 0, 255)
	Ls := clamp((&Vector{float64(I.R), float64(I.G), float64(I.B), 0}).
		MultiplyScalar(r.s.Lights[0].Kspec).
		MultiplyScalar(math.Pow(normal.Dot(H), tex.Shininess)), 0, 255)
	c := clamp(La.Add(Ld).Add(Ls), 0, 255)
	return color.RGBA{uint8(c.X), uint8(c.Y), uint8(c.Z), 255}
}

func clamp(v *Vector, min, max float64) *Vector {
	v.X = math.Min(math.Max(v.X, 0), 255)
	v.Y = math.Min(math.Max(v.Y, 0), 255)
	v.Z = math.Min(math.Max(v.Z, 0), 255)
	return v
}
