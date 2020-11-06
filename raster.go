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
	c Camera // camera
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
		viewMatrix:     IdentityMatrix,
		projMatrix:     IdentityMatrix,
		viewportMatrix: IdentityMatrix,
	}
}

// SetCamera sets the rasterizer camera
func (r *Rasterizer) SetCamera(c Camera) {
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
	r.resetBufs()
	r.initTrans()
	limiter := NewConccurLimiter(runtime.GOMAXPROCS(0))
	for i := 0; i < len(r.s.Objects); i++ {
		o := r.s.Objects[i]
		o.modelMatrix = o.translateMatrix.Mul(o.scaleMatrix)
		o.normalMatrix = o.modelMatrix.Inverse().Transpose()
		for i := 0; i < len(o.triangles); i += int(r.concurrentSize) {
			ii := i
			limiter.Execute(func() {
				for k := int32(0); k < r.concurrentSize; k++ {
					if ii+int(k) >= len(o.triangles) {
						return
					}
					r.draw(o.triangles[ii+int(k)], o.texture, o.modelMatrix, o.normalMatrix)
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

func (r *Rasterizer) draw(tri *Triangle, tex *Texture, modelMatrix, normalMatrix Matrix) {
	v1 := r.VertexShader(tri.v1, modelMatrix)
	v2 := r.VertexShader(tri.v2, modelMatrix)
	v3 := r.VertexShader(tri.v3, modelMatrix)

	// backface culling
	f1 := Vector{v2.Position.X, v2.Position.Y, v2.Position.Z, 1}.Sub(v1.Position)
	f2 := Vector{v3.Position.X, v3.Position.Y, v3.Position.Z, 1}.Sub(v1.Position)
	fN := f1.Cross(f2)
	v := Vector{0, 0, -1, 0}
	if v.Dot(fN) >= 0 {
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
	n1 := normalMatrix.MulVec(tri.v1.Normal)
	n2 := normalMatrix.MulVec(tri.v2.Normal)
	n3 := normalMatrix.MulVec(tri.v3.Normal)
	a := modelMatrix.MulVec(tri.v1.Position)
	b := modelMatrix.MulVec(tri.v2.Position)
	c := modelMatrix.MulVec(tri.v3.Position)

	for x := math.Floor(xMin); x < xMax; x++ {
		for y := math.Floor(yMin); y < yMax; y++ {
			// compute barycentric
			ap := Vector{x, y, 0, 1}.Sub(Vector{v1.Position.X, v1.Position.Y, 0, 1})
			ab := Vector{v2.Position.X, v2.Position.Y, 0, 1}.Sub(Vector{v1.Position.X, v1.Position.Y, 0, 1})
			ac := Vector{v3.Position.X, v3.Position.Y, 0, 1}.Sub(Vector{v1.Position.X, v1.Position.Y, 0, 1})
			bc := Vector{v3.Position.X, v3.Position.Y, 0, 1}.Sub(Vector{v2.Position.X, v2.Position.Y, 0, 1})
			bp := Vector{x, y, 0, 1}.Sub(Vector{v2.Position.X, v2.Position.Y, 0, 1})
			Sabc := ab.Cross(ac).Z
			Sabp := ab.Cross(ap).Z
			Sapc := ap.Cross(ac).Z
			Sbcp := bc.Cross(bp).Z
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
			uv := Vector{
				w1*tri.v1.UV.X + w2*tri.v2.UV.X + w3*tri.v3.UV.X,
				w1*tri.v1.UV.Y + w2*tri.v2.UV.Y + w3*tri.v3.UV.Y,
				0, 1}
			p := Vector{
				w1*a.X + w2*b.X + w3*c.X,
				w1*a.Y + w2*b.Y + w3*c.Y,
				w1*a.Z + w2*b.Z + w3*c.Z, 1}
			n := Vector{
				w1*n1.X + w2*n2.X + w3*n3.X,
				w1*n1.Y + w2*n2.Y + w3*n3.Y,
				w1*n1.Z + w2*n2.Z + w3*n3.Z, 0}.Normalize()
			c := r.FragmentShader(tex, uv, n, p)

			r.lockBuf[idx].Lock()
			r.depthBuf[idx] = z
			r.frameBuf[idx] = c
			r.lockBuf[idx].Unlock()
		}
	}
}

func (r *Rasterizer) resetBufs() {
	size := r.width * r.height
	for i := 0; i < size; i++ {
		// r.frameBuf[i] = color.RGBA{0, 0, 0, 0}
		r.depthBuf[i] = -math.MaxInt64
	}
}

func (r *Rasterizer) initTrans() {
	camPos := r.c.GetPosition()
	camLookAt := r.c.GetLookAt()
	camUp := r.c.GetUp()

	w := (Vector{camLookAt.X, camLookAt.Y, camLookAt.Z, 1}).Sub(camPos).Normalize()
	u := camUp.Cross(w).Mul(-1).Normalize()
	vv := u.Cross(w).Normalize()
	r.viewMatrix = Matrix{
		u.X, u.Y, u.Z, -camPos.X*u.X - camPos.Y*u.Y - camPos.Z*u.Z,
		vv.X, vv.Y, vv.Z, -camPos.X*vv.X - camPos.Y*vv.Y - camPos.Z*vv.Z,
		-w.X, -w.Y, -w.Z, camPos.X*w.X + camPos.Y*w.Y + camPos.Z*w.Z,
		0, 0, 0, 1,
	}
	r.projMatrix = r.c.GetProjectionMatrix()
	r.viewportMatrix = Matrix{
		float64(r.width) / 2, 0, 0, float64(r.width) / 2,
		0, float64(r.height) / 2, 0, float64(r.height) / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}

// VertexShader is a vertex shader that conducts MVP transformations
func (r *Rasterizer) VertexShader(v Vertex, modelMatrix Matrix) Vertex {
	v.Position = v.Position.ApplyMatrix(&modelMatrix).ApplyMatrix(&r.viewMatrix).
		ApplyMatrix(&r.projMatrix).ApplyMatrix(&r.viewportMatrix)
	w := 1.0 / v.Position.W
	v.Position = Vector{v.Position.X * w, v.Position.Y * w, v.Position.Z * w, 1}
	return v
}

// FragmentShader is a Blinn-Phong fragment shader
func (r *Rasterizer) FragmentShader(tex *Texture, uv, normal, p Vector) color.RGBA {
	x := int(math.Floor(uv.X * float64(tex.width)))
	y := int(float64(tex.height) - math.Floor(uv.Y*float64(tex.height)))
	R, G, B, A := tex.data.At(x, y).RGBA()
	I := Vector{float64(R >> 8), float64(G >> 8), float64(B >> 8), float64(A >> 8)}
	L := r.s.Lights[0].Position.Sub(p).Normalize()
	V := r.c.GetPosition().Sub(p).Normalize()
	H := L.Add(V).Normalize()
	nL := normal.Dot(L)
	nH := math.Pow(normal.Dot(H), tex.Shininess)
	blinnPhong := r.s.Lights[0].Kamb + r.s.Lights[0].Kdiff*nL + r.s.Lights[0].Kspec*nH
	c := I.Mul(blinnPhong)
	return color.RGBA{
		uint8(clamp(c.X, 0, 255)),
		uint8(clamp(c.Y, 0, 255)),
		uint8(clamp(c.Z, 0, 255)), 255}
}

func clamp(v, min, max float64) float64 {
	return math.Min(math.Max(v, min), max)
}
