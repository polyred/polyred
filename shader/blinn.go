// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader

import (
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
	"poly.red/texture"
)

var _ Program = &BlinnShader{}

// BlinnShader implements a pair of shader programs that does the
// Blinn-Phong reflectance shading model.
type BlinnShader struct {
	ModelMatrix      math.Mat4
	ViewMatrix       math.Mat4
	ProjectionMatrix math.Mat4
	LightSources     []light.Source
	LightEnviron     []light.Environment
	Kdiff            float32
	Kspec            float32
	Shininess        float32
	Texture          *texture.Texture
}

func (s *BlinnShader) VertexShader(v primitive.Vertex) primitive.Vertex {
	v.Pos = s.ProjectionMatrix.MulM(s.ViewMatrix).MulM(s.ModelMatrix).MulV(v.Pos)
	return v
}

func (s *BlinnShader) FragmentShader(frag primitive.Fragment) color.RGBA {
	x := frag.AttrSmooth["PosModel"].(math.Vec4)
	c := frag.AttrSmooth["PosCam"].(math.Vec4)
	col := frag.Col
	if s.Texture != nil {
		col = s.Texture.Query(0, frag.UV.X, frag.UV.Y)
	}

	LaR := float32(0.0)
	LaG := float32(0.0)
	LaB := float32(0.0)

	for _, e := range s.LightEnviron {
		LaR += e.Intensity() * float32(col.R)
		LaG += e.Intensity() * float32(col.G)
		LaB += e.Intensity() * float32(col.B)
	}

	LdR := float32(0.0)
	LdG := float32(0.0)
	LdB := float32(0.0)

	LsR := float32(0.0)
	LsG := float32(0.0)
	LsB := float32(0.0)

	for _, l := range s.LightSources {
		var (
			L math.Vec4
			I float32
		)
		switch ll := l.(type) {
		case *light.Point:
			Ldir := ll.Position().ToVec4(1).Sub(x)
			L = Ldir.Unit()
			I = ll.Intensity() / Ldir.Len()
		case *light.Directional:
			L = ll.Dir().ToVec4(0).Scale(-1, -1, -1, 1)
			I = ll.Intensity()
		}

		V := c.Sub(x).Unit()
		H := L.Add(V).Unit()
		Ld := math.Clamp(frag.Nor.Dot(L), 0, 1)
		Ls := math.Pow(math.Clamp(frag.Nor.Dot(H), 0, 1), s.Shininess)

		LdR += Ld * float32(col.R) * I
		LdG += Ld * float32(col.G) * I
		LdB += Ld * float32(col.B) * I

		LsR += Ls * float32(l.Color().R) * I
		LsG += Ls * float32(l.Color().G) * I
		LsB += Ls * float32(l.Color().B) * I
	}

	r := LaR + s.Kdiff*LdR + s.Kspec*LsR
	g := LaG + s.Kdiff*LdG + s.Kspec*LsG
	b := LaB + s.Kdiff*LdB + s.Kspec*LsB

	return color.RGBA{
		uint8(math.Clamp(r, 0, 0xff)),
		uint8(math.Clamp(g, 0, 0xff)),
		uint8(math.Clamp(b, 0, 0xff)),
		uint8(math.Clamp(float32(col.A), 0, 0xff))}
}
