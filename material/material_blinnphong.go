// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
	"poly.red/texture"
)

type BlinnPhongMaterial struct {
	tex              *texture.Texture
	kDiff            float32
	kSpec            float32
	shininess        float32
	flatShading      bool
	receiveShadow    bool
	ambientOcclusion bool
}

func (m *BlinnPhongMaterial) Texture() *texture.Texture {
	return m.tex
}

func NewBlinnPhong(opts ...Opt) Material {
	t := &BlinnPhongMaterial{
		tex:              nil,
		kDiff:            0.5,
		kSpec:            1,
		shininess:        1,
		flatShading:      false,
		receiveShadow:    false,
		ambientOcclusion: false,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

func (m *BlinnPhongMaterial) VertexShader(v primitive.Vertex, uniforms map[string]interface{}) primitive.Vertex {
	matModel := uniforms["matModel"].(math.Mat4)
	matView := uniforms["matView"].(math.Mat4)
	matProj := uniforms["matProj"].(math.Mat4)
	matNormal := uniforms["matNormal"].(math.Mat4)

	pos := matProj.MulM(matView).MulM(matModel).MulV(v.Pos)
	return primitive.Vertex{
		Pos: pos,
		Col: v.Col,
		UV:  v.UV,
		Nor: v.Nor.Apply(matNormal),
	}
}

func (m *BlinnPhongMaterial) FragmentShader(col color.RGBA, x, n, fN, c math.Vec4, ls []light.Source, es []light.Environment) color.RGBA {
	LaR := float32(0.0)
	LaG := float32(0.0)
	LaB := float32(0.0)

	for _, e := range es {
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

	if m.flatShading {
		n = fN
	}

	for _, l := range ls {
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
		Ld := math.Clamp(n.Dot(L), 0, 1)
		Ls := math.Pow(math.Clamp(n.Dot(H), 0, 1), m.shininess)

		LdR += Ld * float32(col.R) * I
		LdG += Ld * float32(col.G) * I
		LdB += Ld * float32(col.B) * I

		LsR += Ls * float32(l.Color().R) * I
		LsG += Ls * float32(l.Color().G) * I
		LsB += Ls * float32(l.Color().B) * I
	}

	// The Blinn-Phong Reflection Model
	r := LaR + m.kDiff*LdR + m.kSpec*LsR
	g := LaG + m.kDiff*LdG + m.kSpec*LsG
	b := LaB + m.kDiff*LdB + m.kSpec*LsB

	return color.RGBA{
		uint8(math.Clamp(r, 0, 0xff)),
		uint8(math.Clamp(g, 0, 0xff)),
		uint8(math.Clamp(b, 0, 0xff)),
		uint8(math.Clamp(float32(col.A), 0, 0xff))}
}

func (m *BlinnPhongMaterial) ReceiveShadow() bool {
	return m.receiveShadow
}

func (m *BlinnPhongMaterial) AmbientOcclusion() bool {
	return m.ambientOcclusion
}
