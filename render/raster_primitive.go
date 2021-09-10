// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/shader"
	"poly.red/texture"
)

// DrawPrimitives is a pass that executes Draw call concurrently on all
// given triangle primitives, and draws all geometric and rendering
// information on the given buffer. This primitive uses supplied shader
// programs (i.e. currently supports vertex shader and fragment shader)
//
// See shader.Program for more information regarding shader programming.
func (r *Renderer) DrawPrimitives(
	buf *texture.Buffer, p shader.VertexProgram, idx []uint64, verts []*primitive.Vertex,
) {
	len := len(idx)
	if len%3 != 0 {
		panic("index buffer must be a 3 multiple")
	}

	r.sched.Add(uint64(len / 3))

	for i := 0; i < len; i += 3 {
		vs := verts[i : i+3 : i+3]
		tri := primitive.NewTriangle(vs[0], vs[1], vs[2])
		r.sched.Run(func() {
			if !tri.IsValid() {
				return
			}
			r.DrawPrimitive(buf, p, tri)
		})
	}
	r.sched.Wait()
}

// DrawPrimitive implements a triangle draw call of the rasteriation graphics pipeline.
func (r *Renderer) DrawPrimitive(buf *texture.Buffer, p shader.VertexProgram,
	tri *primitive.Triangle) bool {
	v1 := p(tri.V1)
	v2 := p(tri.V2)
	v3 := p(tri.V3)

	// For perspective corrected interpolation
	recipw := [3]float32{1, 1, 1}
	if _, ok := r.renderCamera.(*camera.Perspective); ok {
		recipw[0] = -1 / v1.Pos.W
		recipw[1] = -1 / v2.Pos.W
		recipw[2] = -1 / v3.Pos.W
	}

	viewportMatrix := math.ViewportMatrix(
		float32(buf.Bounds().Dx()),
		float32(buf.Bounds().Dy()),
	)

	v1.Pos = v1.Pos.Apply(viewportMatrix).Pos()
	v2.Pos = v2.Pos.Apply(viewportMatrix).Pos()
	v3.Pos = v3.Pos.Apply(viewportMatrix).Pos()

	// Back-face culling
	if v2.Pos.Sub(v1.Pos).Cross(v3.Pos.Sub(v1.Pos)).Z < 0 {
		return false
	}

	// View frustum culling
	// TODO: deal with window resizing
	if !r.inViewFrustum(buf, v1.Pos, v2.Pos, v3.Pos) {
		return false
	}

	// All vertices are inside the viewport, let's rasterize directly
	if r.inViewport2(buf, v1.Pos) && r.inViewport2(buf, v2.Pos) && r.inViewport2(buf, v3.Pos) {
		r.rasterize(buf, &v1, &v2, &v3, recipw)
		return true
	}

	// Clipping into smaller triangles
	r.drawClip(buf, &v1, &v2, &v3, recipw)
	return true
}

func (r *Renderer) inViewFrustum(buf *texture.Buffer, v1, v2, v3 math.Vec4) bool {
	// TODO: can be optimize?
	viewportAABB := primitive.NewAABB(
		math.NewVec3(float32(buf.Bounds().Dx()*r.msaa), float32(buf.Bounds().Dy()*r.msaa), 1),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 0, -1),
	)
	triangleAABB := primitive.NewAABB(v1.ToVec3(), v2.ToVec3(), v3.ToVec3())
	return viewportAABB.Intersect(triangleAABB)
}

func (r *Renderer) inViewport2(buf *texture.Buffer, v math.Vec4) bool {
	w := float32(r.msaa * buf.Bounds().Dx())
	h := float32(r.msaa * buf.Bounds().Dy())
	if v.X < 0 || v.X > w || v.Y < 0 || v.Y > h || v.Z > 1 || v.Z < -1 {
		return false
	}
	return true
}

func (r *Renderer) drawClip(buf *texture.Buffer, v1, v2, v3 *primitive.Vertex, recipw [3]float32) {
	w := float32(buf.Bounds().Dx())
	h := float32(buf.Bounds().Dy())

	// Sutherland Hodgman clipping algorithm
	planes := [6]plane{
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
	for _, plane := range planes {
		input := clips
		clips = nil
		if len(input) == 0 {
			clips = nil
			break
		}

		s := input[len(input)-1]
		for _, e := range input {
			if plane.pointInFront(e) {
				if !plane.pointInFront(s) {
					x := plane.intersectSegment(s, e)
					clips = append(clips, x)
				}
				clips = append(clips, e)
			} else if plane.pointInFront(s) {
				x := plane.intersectSegment(s, e)
				clips = append(clips, x)
			}
			s = e
		}
	}

	total := len(clips)
	for i := 2; i < total; i++ {
		b1bc := math.Barycoord(math.NewVec2(clips[0].X, clips[0].Y),
			v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
		b2bc := math.Barycoord(math.NewVec2(clips[i-1].X, clips[i-1].Y),
			v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
		b3bc := math.Barycoord(math.NewVec2(clips[i].X, clips[i].Y),
			v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())

		t1 := primitive.Vertex{
			Pos: math.Vec4{
				X: b1bc[0]*v1.Pos.X + b1bc[1]*v2.Pos.X + b1bc[2]*v3.Pos.X,
				Y: b1bc[0]*v1.Pos.Y + b1bc[1]*v2.Pos.Y + b1bc[2]*v3.Pos.Y,
				Z: b1bc[0]*v1.Pos.Z + b1bc[1]*v2.Pos.Z + b1bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b1bc[0]*v1.UV.X + b1bc[1]*v2.UV.X + b1bc[2]*v3.UV.X,
				Y: b1bc[0]*v1.UV.Y + b1bc[1]*v2.UV.Y + b1bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b1bc[0]*v1.Nor.X + b1bc[1]*v2.Nor.X + b1bc[2]*v3.Nor.X,
				Y: b1bc[0]*v1.Nor.Y + b1bc[1]*v2.Nor.Y + b1bc[2]*v3.Nor.Y,
				Z: b1bc[0]*v1.Nor.Z + b1bc[1]*v2.Nor.Z + b1bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b1bc[0]*float32(v1.Col.R)+b1bc[1]*float32(v2.Col.R)+b1bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b1bc[0]*float32(v1.Col.G)+b1bc[1]*float32(v2.Col.G)+b1bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b1bc[0]*float32(v1.Col.B)+b1bc[1]*float32(v2.Col.B)+b1bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b1bc[0]*float32(v1.Col.A)+b1bc[1]*float32(v2.Col.A)+b1bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}
		t2 := primitive.Vertex{
			Pos: math.Vec4{
				X: b2bc[0]*v1.Pos.X + b2bc[1]*v2.Pos.X + b2bc[2]*v3.Pos.X,
				Y: b2bc[0]*v1.Pos.Y + b2bc[1]*v2.Pos.Y + b2bc[2]*v3.Pos.Y,
				Z: b2bc[0]*v1.Pos.Z + b2bc[1]*v2.Pos.Z + b2bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b2bc[0]*v1.UV.X + b2bc[1]*v2.UV.X + b2bc[2]*v3.UV.X,
				Y: b2bc[0]*v1.UV.Y + b2bc[1]*v2.UV.Y + b2bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b2bc[0]*v1.Nor.X + b2bc[1]*v2.Nor.X + b2bc[2]*v3.Nor.X,
				Y: b2bc[0]*v1.Nor.Y + b2bc[1]*v2.Nor.Y + b2bc[2]*v3.Nor.Y,
				Z: b2bc[0]*v1.Nor.Z + b2bc[1]*v2.Nor.Z + b2bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b2bc[0]*float32(v1.Col.R)+b2bc[1]*float32(v2.Col.R)+b2bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b2bc[0]*float32(v1.Col.G)+b2bc[1]*float32(v2.Col.G)+b2bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b2bc[0]*float32(v1.Col.B)+b2bc[1]*float32(v2.Col.B)+b2bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b2bc[0]*float32(v1.Col.A)+b2bc[1]*float32(v2.Col.A)+b2bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}
		t3 := primitive.Vertex{
			Pos: math.Vec4{
				X: b3bc[0]*v1.Pos.X + b3bc[1]*v2.Pos.X + b3bc[2]*v3.Pos.X,
				Y: b3bc[0]*v1.Pos.Y + b3bc[1]*v2.Pos.Y + b3bc[2]*v3.Pos.Y,
				Z: b3bc[0]*v1.Pos.Z + b3bc[1]*v2.Pos.Z + b3bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b3bc[0]*v1.UV.X + b3bc[1]*v2.UV.X + b3bc[2]*v3.UV.X,
				Y: b3bc[0]*v1.UV.Y + b3bc[1]*v2.UV.Y + b3bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b3bc[0]*v1.Nor.X + b3bc[1]*v2.Nor.X + b3bc[2]*v3.Nor.X,
				Y: b3bc[0]*v1.Nor.Y + b3bc[1]*v2.Nor.Y + b3bc[2]*v3.Nor.Y,
				Z: b3bc[0]*v1.Nor.Z + b3bc[1]*v2.Nor.Z + b3bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b3bc[0]*float32(v1.Col.R)+b3bc[1]*float32(v2.Col.R)+b3bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b3bc[0]*float32(v1.Col.G)+b3bc[1]*float32(v2.Col.G)+b3bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b3bc[0]*float32(v1.Col.B)+b3bc[1]*float32(v2.Col.B)+b3bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b3bc[0]*float32(v1.Col.A)+b3bc[1]*float32(v2.Col.A)+b3bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}

		r.rasterize(buf, &t1, &t2, &t3, recipw)
	}
}

// rasterize implements the rasterization process of a given primitive.
func (r *Renderer) rasterize(buf *texture.Buffer, v1, v2, v3 *primitive.Vertex, recipw [3]float32) {
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
			if !buf.In(x, y) {
				continue
			}

			p := math.NewVec2(float32(x)+0.5, float32(y)+0.5)

			// Compute barycentric coordinates of a triangle in screen
			// space and check if the fragment is inside triangle.
			bc := math.Barycoord(p, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			if bc[0] < -math.Epsilon || bc[1] < -math.Epsilon || bc[2] < -math.Epsilon {
				continue
			}

			// Early Z-test. We normalize depth values to [0, 1], such that
			// the smallest depth value is 0. This collaborate with the buffer
			// clearing.
			z := ((bc[0]*v1.Pos.Z + bc[1]*v2.Pos.Z + bc[2]*v3.Pos.Z) + 1) / 2
			if !buf.DepthTest(x, y, z) {
				continue
			}

			frag := primitive.Fragment{
				X:     x,
				Y:     y,
				Depth: z,
			}

			// Interpolating UV
			uvX := r.interpolate([3]float32{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bc)
			uvY := r.interpolate([3]float32{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bc)
			frag.UV = math.NewVec2(uvX, uvY)

			p1 := math.NewVec2(p.X+1, p.Y)
			p2 := math.NewVec2(p.X, p.Y+1)
			bcx := math.Barycoord(p1, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())
			bcy := math.Barycoord(p2, v1.Pos.ToVec2(), v2.Pos.ToVec2(), v3.Pos.ToVec2())

			uvdU := r.interpolate([3]float32{v1.UV.X, v2.UV.X, v3.UV.X}, recipw, bcx)
			uvdX := r.interpolate([3]float32{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcx)
			uvdV := r.interpolate([3]float32{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			uvdY := r.interpolate([3]float32{v1.UV.Y, v2.UV.Y, v3.UV.Y}, recipw, bcy)
			frag.Du = math.Sqrt((uvdU-uvX)*(uvdU-uvX) + (uvdX-uvY)*(uvdX-uvY))
			frag.Dv = math.Sqrt((uvdV-uvX)*(uvdV-uvX) + (uvdY-uvY)*(uvdY-uvY))

			// Interpolate normal
			if !v1.Nor.IsZero() && !v2.Nor.IsZero() && !v3.Nor.IsZero() {
				nx := r.interpolate([3]float32{v1.Nor.X, v2.Nor.X, v3.Nor.X}, recipw, bc)
				ny := r.interpolate([3]float32{v1.Nor.Y, v2.Nor.Y, v3.Nor.Y}, recipw, bc)
				nz := r.interpolate([3]float32{v1.Nor.Z, v2.Nor.Z, v3.Nor.Z}, recipw, bc)
				frag.Nor = math.NewVec4(nx, ny, nz, 0)
			}

			// Interpolate colors
			if v1.Col != color.Discard || v2.Col != color.Discard || v3.Col != color.Discard {
				cr := r.interpolate([3]float32{float32(v1.Col.R), float32(v2.Col.R), float32(v3.Col.R)}, recipw, bc)
				cg := r.interpolate([3]float32{float32(v1.Col.G), float32(v2.Col.G), float32(v3.Col.G)}, recipw, bc)
				cb := r.interpolate([3]float32{float32(v1.Col.B), float32(v2.Col.B), float32(v3.Col.B)}, recipw, bc)
				ca := r.interpolate([3]float32{float32(v1.Col.A), float32(v2.Col.A), float32(v3.Col.A)}, recipw, bc)
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

			buf.Set(x, y, texture.Fragment{
				Ok:       true,
				Fragment: frag,
			})
		}
	}
}

// interpoVaryings perspective correct interpolates
func (r *Renderer) interpoVaryings(v1, v2, v3, frag map[string]interface{},
	recipw, bc [3]float32) {
	l := len(frag)
	for name := range v1 {
		switch val1 := v1[name].(type) {
		case float32:
			if l > 0 {
				frag[name] = r.interpolate([3]float32{
					val1, v2[name].(float32), v3[name].(float32),
				}, recipw, bc)
			}
		case math.Vec2:
			x := r.interpolate([3]float32{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float32{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec2(x, y)
			}
		case math.Vec3:
			x := r.interpolate([3]float32{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float32{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float32{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			if l > 0 {
				frag[name] = math.NewVec3(x, y, z)
			}
		case math.Vec4:
			x := r.interpolate([3]float32{
				val1.X,
				v2[name].(math.Vec4).X,
				v3[name].(math.Vec4).X,
			}, recipw, bc)
			y := r.interpolate([3]float32{
				val1.Y,
				v2[name].(math.Vec4).Y,
				v3[name].(math.Vec4).Y,
			}, recipw, bc)
			z := r.interpolate([3]float32{
				val1.Z,
				v2[name].(math.Vec4).Z,
				v3[name].(math.Vec4).Z,
			}, recipw, bc)
			w := r.interpolate([3]float32{
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
func (r *Renderer) interpolate(varying, recipw, barycoord [3]float32) float32 {
	recipw[0] *= barycoord[0]
	recipw[1] *= barycoord[1]
	recipw[2] *= barycoord[2]
	norm := recipw[0]*varying[0] + recipw[1]*varying[1] + recipw[2]*varying[2]
	if r.renderPerspect {
		norm *= 1 / (recipw[0] + recipw[1] + recipw[2])
	}
	return norm
}
