// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
	"poly.red/shader"
)

type BlinnPhongMaterial struct {
	tex              *buffer.Texture
	kDiff            float32
	kSpec            float32
	shininess        float32
	flatShading      bool
	receiveShadow    bool
	ambientOcclusion bool
}

func (m *BlinnPhongMaterial) Texture() *buffer.Texture {
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

func (m *BlinnPhongMaterial) VertexShader(v *primitive.Vertex) *primitive.Vertex {
	mvp := v.AttrFlat[shader.MVPAttr].(*shader.MVP)
	pos := mvp.Proj.MulM(mvp.View).MulM(mvp.Model).MulV(v.Pos)
	vv := primitive.NewVertex(
		primitive.Pos(pos),
		primitive.Col(v.Col),
		primitive.UV(v.UV),
		primitive.Nor(v.Nor.Apply(mvp.Normal)),
	)
	vv.AttrFlat[shader.MVPAttr] = mvp
	return vv
}

func (m *BlinnPhongMaterial) FragmentShader(
	info buffer.Fragment, c math.Vec3[float32],
	ls []light.Source, es []light.Environment,
) color.RGBA {
	lod := float32(0.0)
	if m.tex.UseMipmap() {
		siz := float32(m.tex.Size()) * math.Sqrt(math.Max(info.Du, info.Dv))
		if siz < 1 {
			siz = 1
		}
		lod = math.Log2(siz)
	}
	col := m.tex.Query(lod, info.U, 1-info.V)

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

	n := info.Nor
	if m.flatShading {
		n = info.AttrFlat["fN"].(math.Vec4[float32])
	}
	x := info.AttrFlat["Pos"].(math.Vec4[float32])

	for _, l := range ls {
		var (
			L math.Vec4[float32]
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

		V := c.ToVec4(1).Sub(x).Unit()
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
