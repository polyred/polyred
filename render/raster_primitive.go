// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/shader"
	"poly.red/texture/buffer"
)

// DrawPrimitives is a pass that executes Draw call concurrently on all
// given triangle primitives, and draws all geometric and rendering
// information on the given buffer. This primitive uses supplied shader
// programs (i.e. currently supports vertex shader and fragment shader)
//
// See shader.Program for more information regarding shader programming.
func (r *Renderer) DrawPrimitives(m geometry.Renderable, p shader.VertexProgram) {
	idx := m.IndexBuffer()
	verts := m.VertexBuffer()

	len := len(idx)
	if len%3 != 0 {
		panic("index buffer must be a 3 multiple")
	}

	r.sched.Add(uint64(len / 3))
	for i := 0; i < len; i += 3 {
		v1 := verts[i]
		v2 := verts[i+1]
		v3 := verts[i+2]
		r.sched.Run(func() {
			if primitive.IsValidTriangle(v1.Pos.ToVec3(), v2.Pos.ToVec3(), v3.Pos.ToVec3()) {
				r.DrawPrimitive(p, v1, v2, v3)
			}
		})
	}
	r.sched.Wait()
}

// DrawPrimitive implements a triangle draw call of the rasteriation graphics pipeline.
func (r *Renderer) DrawPrimitive(p shader.VertexProgram, p1, p2, p3 *primitive.Vertex) {
	v1 := p(*p1)
	v2 := p(*p2)
	v3 := p(*p3)

	// For perspective corrected interpolation
	recipw := [3]float64{1, 1, 1}
	if r.renderPerspect {
		recipw[0] = -1 / v1.Pos.W
		recipw[1] = -1 / v2.Pos.W
		recipw[2] = -1 / v3.Pos.W
	}
	r.rasterize(&v1, &v2, &v3, recipw)

	// v1ok := isVisible(v1.Pos)
	// v2ok := isVisible(v2.Pos)
	// v3ok := isVisible(v3.Pos)
	// if v1ok && v2ok && v3ok {
	// 	r.rasterize(buf, &v1, &v2, &v3, recipw)
	// 	return
	// }
	// r.clipping(buf, &v1, &v2, &v3, recipw)
}

func isVisible(v math.Vec4) bool {
	return (math.Abs(v.X) <= -v.W) && (math.Abs(v.Y) <= -v.W) && (math.Abs(v.Z) <= -v.W)
}

// clipping clips the given triangle into smaller triangles then rasterize
// then onto the given buffer.
func (r *Renderer) clipping(v1, v2, v3 *primitive.Vertex, recipw [3]float64) {
	w := float64(r.buf.Bounds().Dx())
	h := float64(r.buf.Bounds().Dy())

	// Sutherland Hodgman clipping algorithm
	planes := [6]struct{ pos, nor math.Vec4 }{
		{math.NewVec4(w, 0, 0, 1), math.NewVec4(-1, 0, 0, 1)},
		{math.NewVec4(0, 0, 0, 1), math.NewVec4(1, 0, 0, 1)},
		{math.NewVec4(0, h, 0, 1), math.NewVec4(0, -1, 0, 1)},
		{math.NewVec4(0, 0, 0, 1), math.NewVec4(0, 1, 0, 1)},
		{math.NewVec4(0, 0, 1, 1), math.NewVec4(0, 0, -1, 1)},
		{math.NewVec4(0, 0, -1, 1), math.NewVec4(0, 0, 1, 1)},
	}

	// TODO: need optimize
	input := []math.Vec4{v1.Pos, v2.Pos, v3.Pos}
	clips := input
	for i := 0; i < 6; i++ {
		input := clips
		clips = nil
		if len(input) == 0 {
			clips = nil
			break
		}

		// fmt.Println("ok")

		s := input[len(input)-1]
		for _, e := range input {
			if e.Sub(planes[i].pos).Dot(planes[i].nor) > 0 {
				if !(s.Sub(planes[i].pos).Dot(planes[i].nor) > 0) {
					u := e.Sub(s)
					w := s.Sub(planes[i].pos)
					d := planes[i].nor.Dot(u)
					n := -planes[i].nor.Dot(w)
					ss := n / d
					x := s.Add(u.Scale(ss, ss, ss, ss))
					clips = append(clips, x)
				}
				clips = append(clips, e)
			} else if s.Sub(planes[i].pos).Dot(planes[i].nor) > 0 {
				u := e.Sub(s)
				w := s.Sub(planes[i].pos)
				d := planes[i].nor.Dot(u)
				n := -planes[i].nor.Dot(w)
				ss := n / d
				x := s.Add(u.Scale(ss, ss, ss, ss))
				clips = append(clips, x)
			}
			s = e
		}
	}

	l := len(clips)
	v1p := v1.Pos.ToVec2()
	v2p := v2.Pos.ToVec2()
	v3p := v3.Pos.ToVec2()

	for i := 2; i < l; i++ {
		b1bc := math.Barycoord(clips[0].ToVec2(), v1p, v2p, v3p)
		t1 := primitive.Vertex{
			Pos: math.Vec4{
				X: b1bc[0]*v1.Pos.X + b1bc[1]*v2.Pos.X + b1bc[2]*v3.Pos.X,
				Y: b1bc[0]*v1.Pos.Y + b1bc[1]*v2.Pos.Y + b1bc[2]*v3.Pos.Y,
				Z: b1bc[0]*v1.Pos.Z + b1bc[1]*v2.Pos.Z + b1bc[2]*v3.Pos.Z,
				W: b1bc[0]*v1.Pos.W + b1bc[1]*v2.Pos.W + b1bc[2]*v3.Pos.W,
			},
		}
		u := r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, b1bc)
		v := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, b1bc)
		t1.UV = math.NewVec4(u, v, 0, 1)
		if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
			nx := r.interpolate([3]float64{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, b1bc)
			ny := r.interpolate([3]float64{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, b1bc)
			nz := r.interpolate([3]float64{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, b1bc)
			t1.Nor = math.NewVec4(nx, ny, nz, 0)
		}
		if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
			cr := r.interpolate([3]float64{float64(v1.Col.R), float64(v2.Col.R), float64(v3.Col.R)}, recipw, b1bc)
			cg := r.interpolate([3]float64{float64(v1.Col.G), float64(v2.Col.G), float64(v3.Col.G)}, recipw, b1bc)
			cb := r.interpolate([3]float64{float64(v1.Col.B), float64(v2.Col.B), float64(v3.Col.B)}, recipw, b1bc)
			ca := r.interpolate([3]float64{float64(v1.Col.A), float64(v2.Col.A), float64(v3.Col.A)}, recipw, b1bc)
			t1.Col = color.RGBA{
				R: uint8(math.Clamp(cr, 0, 0xff)),
				G: uint8(math.Clamp(cg, 0, 0xff)),
				B: uint8(math.Clamp(cb, 0, 0xff)),
				A: uint8(math.Clamp(ca, 0, 0xff)),
			}
		}

		b2bc := math.Barycoord(clips[i-1].ToVec2(), v1p, v2p, v3p)
		t2 := primitive.Vertex{
			Pos: math.Vec4{
				X: b2bc[0]*v1.Pos.X + b2bc[1]*v2.Pos.X + b2bc[2]*v3.Pos.X,
				Y: b2bc[0]*v1.Pos.Y + b2bc[1]*v2.Pos.Y + b2bc[2]*v3.Pos.Y,
				Z: b2bc[0]*v1.Pos.Z + b2bc[1]*v2.Pos.Z + b2bc[2]*v3.Pos.Z,
				W: b2bc[0]*v1.Pos.W + b2bc[1]*v2.Pos.W + b2bc[2]*v3.Pos.W,
			},
		}
		u = r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, b2bc)
		v = r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, b2bc)
		t2.UV = math.NewVec4(u, v, 0, 1)
		if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
			nx := r.interpolate([3]float64{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, b2bc)
			ny := r.interpolate([3]float64{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, b2bc)
			nz := r.interpolate([3]float64{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, b2bc)
			t2.Nor = math.NewVec4(nx, ny, nz, 0)
		}
		if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
			cr := r.interpolate([3]float64{float64(v1.Col.R), float64(v2.Col.R), float64(v3.Col.R)}, recipw, b2bc)
			cg := r.interpolate([3]float64{float64(v1.Col.G), float64(v2.Col.G), float64(v3.Col.G)}, recipw, b2bc)
			cb := r.interpolate([3]float64{float64(v1.Col.B), float64(v2.Col.B), float64(v3.Col.B)}, recipw, b2bc)
			ca := r.interpolate([3]float64{float64(v1.Col.A), float64(v2.Col.A), float64(v3.Col.A)}, recipw, b2bc)
			t1.Col = color.RGBA{
				R: uint8(math.Clamp(cr, 0, 0xff)),
				G: uint8(math.Clamp(cg, 0, 0xff)),
				B: uint8(math.Clamp(cb, 0, 0xff)),
				A: uint8(math.Clamp(ca, 0, 0xff)),
			}
		}

		b3bc := math.Barycoord(clips[i].ToVec2(), v1p, v2p, v3p)
		t3 := primitive.Vertex{
			Pos: math.Vec4{
				X: b3bc[0]*v1.Pos.X + b3bc[1]*v2.Pos.X + b3bc[2]*v3.Pos.X,
				Y: b3bc[0]*v1.Pos.Y + b3bc[1]*v2.Pos.Y + b3bc[2]*v3.Pos.Y,
				Z: b3bc[0]*v1.Pos.Z + b3bc[1]*v2.Pos.Z + b3bc[2]*v3.Pos.Z,
				W: b3bc[0]*v1.Pos.W + b3bc[1]*v2.Pos.W + b3bc[2]*v3.Pos.W,
			},
		}
		u = r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, b3bc)
		v = r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, b3bc)
		t3.UV = math.NewVec4(u, v, 0, 1)
		if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
			nx := r.interpolate([3]float64{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, b3bc)
			ny := r.interpolate([3]float64{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, b3bc)
			nz := r.interpolate([3]float64{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, b3bc)
			t3.Nor = math.NewVec4(nx, ny, nz, 0)
		}
		if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
			cr := r.interpolate([3]float64{float64(v1.Col.R), float64(v2.Col.R), float64(v3.Col.R)}, recipw, b3bc)
			cg := r.interpolate([3]float64{float64(v1.Col.G), float64(v2.Col.G), float64(v3.Col.G)}, recipw, b3bc)
			cb := r.interpolate([3]float64{float64(v1.Col.B), float64(v2.Col.B), float64(v3.Col.B)}, recipw, b3bc)
			ca := r.interpolate([3]float64{float64(v1.Col.A), float64(v2.Col.A), float64(v3.Col.A)}, recipw, b3bc)
			t1.Col = color.RGBA{
				R: uint8(math.Clamp(cr, 0, 0xff)),
				G: uint8(math.Clamp(cg, 0, 0xff)),
				B: uint8(math.Clamp(cb, 0, 0xff)),
				A: uint8(math.Clamp(ca, 0, 0xff)),
			}
		}
		r.rasterize(&t1, &t2, &t3, recipw)
	}
}

// rasterize implements the rasterization process of a given primitive.
func (r *Renderer) rasterize(v1, v2, v3 *primitive.Vertex, recipw [3]float64) {
	v1.Pos = v1.Pos.Apply(r.viewportMatrix).Pos()
	v2.Pos = v2.Pos.Apply(r.viewportMatrix).Pos()
	v3.Pos = v3.Pos.Apply(r.viewportMatrix).Pos()

	// TODO: which should be the first?

	// Back-face culling
	if v2.Pos.Sub(v1.Pos).Cross(v3.Pos.Sub(v1.Pos)).Z < 0 {
		return
	}

	// View frustum culling
	if !r.cullViewFrustum(v1.Pos, v2.Pos, v3.Pos) {
		return
	}

	// Compute AABB make the AABB a little bigger that align with
	// pixels to contain the entire triangle
	aabb := primitive.NewAABB(v1.Pos.ToVec3(), v2.Pos.ToVec3(), v3.Pos.ToVec3())
	xmin := int(math.Round(aabb.Min.X) - 1)
	xmax := int(math.Round(aabb.Max.X) + 1)
	ymin := int(math.Round(aabb.Min.Y) - 1)
	ymax := int(math.Round(aabb.Max.Y) + 1)

	// TODO: parallize this loop, smarter scheduling to minimize lock
	// contention
	for x := xmin; x <= xmax; x++ {
		for y := ymin; y <= ymax; y++ {
			if !r.buf.In(x, y) {
				continue
			}

			p := math.NewVec2(float64(x)+0.5, float64(y)+0.5)

			// Compute barycentric coordinates of a triangle in screen
			// space and check if the fragment is inside triangle.
			bc := math.Barycoord(p, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			if bc[0] < -math.Epsilon ||
				bc[1] < -math.Epsilon ||
				bc[2] < -math.Epsilon {
				continue
			}

			// Early Z-test. We normalize depth values to [0, 1], such that
			// the smallest depth value is 0. This collaborate with the buffer
			// clearing.
			z := ((bc[0]*v1.Pos.Z + bc[1]*v2.Pos.Z + bc[2]*v3.Pos.Z) + 1) / 2
			if !r.buf.DepthTest(x, y, z) {
				continue
			}

			frag := primitive.Fragment{
				X:     x,
				Y:     y,
				Depth: z,
			}

			// Interpolating UV
			uvX := r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bc)
			uvY := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bc)
			frag.UV = math.NewVec2(uvX, uvY)

			p1 := math.NewVec2(p.X+1, p.Y)
			p2 := math.NewVec2(p.X, p.Y+1)
			bcx := math.Barycoord(p1, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			bcy := math.Barycoord(p2, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())

			uvdU := r.interpolate([3]float64{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bcx)
			uvdX := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcx)
			uvdV := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			uvdY := r.interpolate([3]float64{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			frag.Du = math.Sqrt((uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY))
			frag.Dv = math.Sqrt((uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY))

			// Interpolate normal
			if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
				nx := r.interpolate([3]float64{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, bc)
				ny := r.interpolate([3]float64{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, bc)
				nz := r.interpolate([3]float64{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, bc)
				frag.Nor = math.NewVec4(nx, ny, nz, 0)
			}

			// Interpolate colors
			if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
				cr := r.interpolate([3]float64{float64(v1.Col.R), float64(v2.Col.R), float64(v3.Col.R)}, recipw, bc)
				cg := r.interpolate([3]float64{float64(v1.Col.G), float64(v2.Col.G), float64(v3.Col.G)}, recipw, bc)
				cb := r.interpolate([3]float64{float64(v1.Col.B), float64(v2.Col.B), float64(v3.Col.B)}, recipw, bc)
				ca := r.interpolate([3]float64{float64(v1.Col.A), float64(v2.Col.A), float64(v3.Col.A)}, recipw, bc)
				frag.Col = color.RGBA{
					R: uint8(math.Clamp(cr, 0, 0xff)),
					G: uint8(math.Clamp(cg, 0, 0xff)),
					B: uint8(math.Clamp(cb, 0, 0xff)),
					A: uint8(math.Clamp(ca, 0, 0xff)),
				}
			}

			// Interpolate custom varying
			if len(v1.AttrSmooth) > 0 {
				r.interpoVaryings(v1.AttrSmooth, v2.AttrSmooth, v3.AttrSmooth, frag.AttrSmooth, recipw, bc)
			}

			r.buf.Set(x, y, buffer.Fragment{
				Ok:       true,
				Fragment: frag,
			})
		}
	}
}

// ~200ns
func (r *Renderer) cullViewFrustum(v1, v2, v3 math.Vec4) bool {
	// TODO: can be optimize?
	viewportAABB := primitive.NewAABB(
		math.NewVec3(float64(r.buf.Bounds().Dx()), float64(r.buf.Bounds().Dy()), 1),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 0, -1),
	)
	triangleAABB := primitive.NewAABB(v1.ToVec3(), v2.ToVec3(), v3.ToVec3())
	return viewportAABB.Intersect(triangleAABB)
}

// interpoVaryings perspective correct interpolates
func (r *Renderer) interpoVaryings(v1, v2, v3, frag map[string]interface{},
	recipw, bc [3]float64) {
	l := len(frag)
	for name := range v1 {
		switch val1 := v1[name].(type) {
		case float64:
			if l > 0 {
				frag[name] = r.interpolate([3]float64{
					val1, v2[name].(float64), v3[name].(float64),
				}, recipw, bc)
			}
		case math.Vec2:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec2(x, y)
			}
		case math.Vec3:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float64{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec3(x, y, z)
			}
		case math.Vec4:
			x := r.interpolate([3]float64{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float64{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float64{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			w := r.interpolate([3]float64{
				val1.W,
				v2[name].(math.Vec4).W,
				v3[name].(math.Vec4).W,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec4(x, y, z, w)
			}
		}
	}
}

// interpolate does perspective-correct interpolation for the given varying.
//
// See: Low, Kok-Lim. "Perspective-correct interpolation." Technical writing,
// Department of Computer Science, University of North Carolina at
// Chapel Hill (2002).
func (r *Renderer) interpolate(varying, recipw, barycoord [3]float64) float64 {
	recipw[0] *= barycoord[0]
	recipw[1] *= barycoord[1]
	recipw[2] *= barycoord[2]
	norm := recipw[0]*varying[0] + recipw[1]*varying[1] + recipw[2]*varying[2]
	if r.renderPerspect {
		norm *= 1 / (recipw[0] + recipw[1] + recipw[2])
	}
	return norm
}
