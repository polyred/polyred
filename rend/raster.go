// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package rend

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"sync"
	"sync/atomic"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/utils"
)

// Rasterizer is a CPU rasterizer
type Rasterizer struct {
	width  int
	height int
	msaa   int
	s      *Scene

	lockBuf        []sync.Mutex
	concurrentSize int32 // atomic
}

// NewRasterizer creates a new rasterizer
func NewRasterizer(width, height, msaa int) *Rasterizer {
	return &Rasterizer{
		width:          width * msaa,
		height:         height * msaa,
		msaa:           msaa,
		lockBuf:        make([]sync.Mutex, width*height*msaa),
		concurrentSize: 128, // empirical, see benchmark
	}
}

// Render renders a scene.
func (r *Rasterizer) Render(s *Scene) []color.RGBA {
	r.s = s
	frameBuf := make([]color.RGBA, r.width*r.height)
	depthBuf := make([]float64, r.width*r.height)
	size := r.width * r.height
	for i := 0; i < size; i++ {
		frameBuf[i] = color.RGBA{0, 0, 0, 0}
		depthBuf[i] = -1
	}

	limiter := utils.NewConccurLimiter(runtime.GOMAXPROCS(0))

	for m := 0; m < len(s.Meshes); m++ {
		mesh := s.Meshes[m]

		uniforms := map[string]math.Matrix{
			"matModel":  mesh.ModelMatrix(),
			"matView":   s.Camera.ViewMatrix(),
			"matProj":   s.Camera.ProjMatrix(),
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
					r.draw(frameBuf, depthBuf, uniforms, mesh.Faces[ii+int(k)], mesh.Material)
				}
			})
		}
	}

	limiter.Wait()
	return frameBuf
}

func (r *Rasterizer) barycoord(x, y int, t1, t2, t3 math.Vector) (w1, w2, w3 float64) {
	ap := math.NewVector(float64(x)-t1.X, float64(y)-t1.Y, 0, 0)
	ab := math.NewVector(t2.X-t1.X, t2.Y-t1.Y, 0, 0)
	ac := math.NewVector(t3.X-t1.X, t3.Y-t1.Y, 0, 0)
	bc := math.NewVector(t3.X-t2.X, t3.Y-t2.Y, 0, 0)
	bp := math.NewVector(float64(x)-t2.X, float64(y)-t2.Y, 0, 0)
	Sabc := ab.Cross(ac).Z
	Sabp := ab.Cross(ap).Z
	Sapc := ap.Cross(ac).Z
	Sbcp := bc.Cross(bp).Z
	w1, w2, w3 = Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc
	return
}

func (r *Rasterizer) draw(frameBuf []color.RGBA, depthBuf []float64, uniforms map[string]math.Matrix, tri *geometry.Triangle, mat material.Material) {
	matModel := uniforms["matModel"]
	m1 := tri.V1.Position.Apply(matModel)
	m2 := tri.V1.Position.Apply(matModel)
	m3 := tri.V1.Position.Apply(matModel)

	t1 := r.vertexShader(tri.V1, uniforms)
	t2 := r.vertexShader(tri.V2, uniforms)
	t3 := r.vertexShader(tri.V3, uniforms)

	if r.isBackFace(t1.Position, t2.Position, t3.Position) {
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
			if z <= depthBuf[idx] {
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
			uv := math.NewVector(
				(w1*t1.UV.X+w2*t2.UV.X+w3*t3.UV.X)/Z,
				(w1*t1.UV.Y+w2*t2.UV.Y+w3*t3.UV.Y)/Z,
				0,
				1,
			)

			// Compute du dv
			w1x, w2x, w3x := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
			w1y, w2y, w3y := r.barycoord(x+1, y, t1.Position, t2.Position, t3.Position)
			uvX := math.NewVector(
				(w1x*t1.UV.X+w2x*t2.UV.X+w3x*t3.UV.X)/Z,
				(w1x*t1.UV.Y+w2x*t2.UV.Y+w3x*t3.UV.Y)/Z,
				0,
				1,
			)
			uvY := math.NewVector(
				(w1y*t1.UV.X+w2y*t2.UV.X+w3y*t3.UV.X)/Z,
				(w1y*t1.UV.Y+w2y*t2.UV.Y+w3y*t3.UV.Y)/Z,
				0,
				1,
			)
			lod := math.Log2(
				math.Max(uvX.Sub(uv).Len(), uvY.Sub(uv).Len()))
			col := mat.Texture().Query(uv.X, 1-uv.Y, lod)

			// normal interpolation
			n := math.NewVector(
				(w1*t1.Normal.X+w2*t2.Normal.X+w3*t3.Normal.X)/Z,
				(w1*t1.Normal.Y+w2*t2.Normal.Y+w3*t3.Normal.Y)/Z,
				(w1*t1.Normal.Z+w2*t2.Normal.Z+w3*t3.Normal.Z)/Z,
				0,
			)
			pos := math.NewVector(
				(w1*m1.X+w2*m1.X+w3*m1.X)/Z,
				(w1*m2.Y+w2*m2.Y+w3*m2.Y)/Z,
				(w1*m3.Z+w2*m3.Z+w3*m3.Z)/Z,
				1,
			)
			col = mat.Shader(col, pos, n, r.s.Lights[0].Position(), r.s.Camera.Position())

			r.fragmentProcessing(frameBuf, depthBuf, x, y, z, col)
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

func (r *Rasterizer) isBackFace(v1, v2, v3 math.Vector) bool {
	fN := v2.Sub(v1).Cross(v3.Sub(v1))
	return math.NewVector(0, 0, -1, 0).Dot(fN) >= 0
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

func (r *Rasterizer) fragmentProcessing(frameBuf []color.RGBA, depthBuf []float64, x, y int, z float64, col color.RGBA) {
	if x < 0 || y >= r.width {
		return
	}
	if x < 0 || y >= r.height {
		return
	}

	idx := x + y*r.width
	r.lockBuf[idx].Lock()
	frameBuf[idx] = col
	depthBuf[idx] = z
	r.lockBuf[idx].Unlock()
}

// SetConcurrencySize sets the number of triangles that is processed in parallel
func (r *Rasterizer) SetConcurrencySize(new int32) (old int32) {
	old = atomic.SwapInt32(&r.concurrentSize, new)
	return
}

// // Render renders a scene graph
// func (r *Rasterizer) Render() {
// 	r.resetBufs()
// 	r.initTrans()
//
// 	for i := 0; i < len(r.s.Objects); i++ {
// 		o := r.s.Objects[i]
// 		o.modelMatrix = o.translateMatrix.Mul(o.scaleMatrix)
// 		o.normalMatrix = o.modelMatrix.Inverse().Transpose()
// 		for i := 0; i < len(o.triangles); i += int(r.concurrentSize) {
// 			ii := i
// 				for k := int32(0); k < r.concurrentSize; k++ {
// 					if ii+int(k) >= len(o.triangles) {
// 						return
// 					}
// 					r.draw(o.triangles[ii+int(k)], o.texture, o.modelMatrix, o.normalMatrix)
// 				}
// 			})
// 		}
// 	}
// }

// Save stores the current frame buffer to a newly created file.
func (r *Rasterizer) Save(frameBuf []color.RGBA, filename string) {
	err := r.flushFrameBuffer(frameBuf, filename)
	if err != nil {
		panic(fmt.Errorf("cannot save the render result, err: %v", err))
	}
}

// flushFrameBuffer writes the frame buffer to an image
func (r *Rasterizer) flushFrameBuffer(frameBuf []color.RGBA, filename string) error {
	m := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{r.width, r.height}})
	for i := 0; i < r.width; i++ {
		for j := 0; j < r.height; j++ {
			m.Set(i, r.height-j, frameBuf[j*r.width+i])
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

// func (r *Rasterizer) draw(tri *Triangle, tex *Texture, modelMatrix, normalMatrix Matrix) {
// 	v1 := r.VertexShader(tri.v1, modelMatrix)
// 	v2 := r.VertexShader(tri.v2, modelMatrix)
// 	v3 := r.VertexShader(tri.v3, modelMatrix)

// 	// backface culling
// 	f1 := math.Vector{v2.Position.X, v2.Position.Y, v2.Position.Z, 1}.Sub(v1.Position)
// 	f2 := math.Vector{v3.Position.X, v3.Position.Y, v3.Position.Z, 1}.Sub(v1.Position)
// 	fN := f1.Cross(f2)
// 	v := math.Vector{0, 0, -1, 0}
// 	if v.Dot(fN) >= 0 {
// 		return
// 	}

// 	box := NewAABB(v1, v2, v3)

// 	// view frustum culling
// 	xMax := math.Min(box.max.X, float64(r.width))
// 	xMin := math.Max(box.min.X, 0)
// 	yMax := math.Min(box.max.Y, float64(r.height))
// 	yMin := math.Max(box.min.Y, 0)
// 	if xMin > xMax && yMin > yMax {
// 		return
// 	}

// 	// compute normals and shading point in world space for fragment shading
// 	n1 := normalMatrix.MulVec(tri.v1.Normal)
// 	n2 := normalMatrix.MulVec(tri.v2.Normal)
// 	n3 := normalMatrix.MulVec(tri.v3.Normal)
// 	a := modelMatrix.MulVec(tri.v1.Position)
// 	b := modelMatrix.MulVec(tri.v2.Position)
// 	c := modelMatrix.MulVec(tri.v3.Position)

// 	for x := math.Floor(xMin); x < xMax; x++ {
// 		for y := math.Floor(yMin); y < yMax; y++ {
// 			// compute barycentric
// 			ap := math.Vector{x, y, 0, 1}.Sub(math.Vector{v1.Position.X, v1.Position.Y, 0, 1})
// 			ab := math.Vector{v2.Position.X, v2.Position.Y, 0, 1}.Sub(math.Vector{v1.Position.X, v1.Position.Y, 0, 1})
// 			ac := math.Vector{v3.Position.X, v3.Position.Y, 0, 1}.Sub(math.Vector{v1.Position.X, v1.Position.Y, 0, 1})
// 			bc := math.Vector{v3.Position.X, v3.Position.Y, 0, 1}.Sub(math.Vector{v2.Position.X, v2.Position.Y, 0, 1})
// 			bp := math.Vector{x, y, 0, 1}.Sub(math.Vector{v2.Position.X, v2.Position.Y, 0, 1})
// 			Sabc := ab.Cross(ac).Z
// 			Sabp := ab.Cross(ap).Z
// 			Sapc := ap.Cross(ac).Z
// 			Sbcp := bc.Cross(bp).Z
// 			w1, w2, w3 := Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc

// 			// skip frags outside the triangle
// 			if w1 < 0 || w2 < 0 || w3 < 0 {
// 				continue
// 			}

// 			z := w1*v1.Position.Z + w2*v2.Position.Z + w3*v3.Position.Z
// 			idx := int(y*float64(r.width) + x)
// 			r.lockBuf[idx].Lock()
// 			if z < r.depthBuf[idx] {
// 				r.lockBuf[idx].Unlock()
// 				continue
// 			}
// 			r.lockBuf[idx].Unlock()

// 			// uv interpolation
// 			uv := math.Vector{
// 				w1*tri.v1.UV.X + w2*tri.v2.UV.X + w3*tri.v3.UV.X,
// 				w1*tri.v1.UV.Y + w2*tri.v2.UV.Y + w3*tri.v3.UV.Y,
// 				0, 1}
// 			p := math.Vector{
// 				w1*a.X + w2*b.X + w3*c.X,
// 				w1*a.Y + w2*b.Y + w3*c.Y,
// 				w1*a.Z + w2*b.Z + w3*c.Z, 1}
// 			n := math.Vector{
// 				w1*n1.X + w2*n2.X + w3*n3.X,
// 				w1*n1.Y + w2*n2.Y + w3*n3.Y,
// 				w1*n1.Z + w2*n2.Z + w3*n3.Z, 0}.Normalize()
// 			c := r.FragmentShader(tex, uv, n, p)

// 			r.lockBuf[idx].Lock()
// 			r.depthBuf[idx] = z
// 			r.frameBuf[idx] = c
// 			r.lockBuf[idx].Unlock()
// 		}
// 	}
// }
