// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/math"
)

type BlinnPhongMaterial struct {
	tex              *image.Texture
	kDiff            float64
	kSpec            float64
	shininess        float64
	flatShading      bool
	receiveShadow    bool
	ambientOcclusion bool
}

func (m *BlinnPhongMaterial) Texture() *image.Texture {
	return m.tex
}

type BlinnPhongMaterialOption func(m *BlinnPhongMaterial)

func WithBlinnPhongTexture(tex *image.Texture) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.tex = tex
	}
}

func WithBlinnPhongFactors(Kdiff, Kspec float64) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.kDiff = Kdiff
		m.kSpec = Kspec
	}
}

func WithBlinnPhongShininess(p float64) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.shininess = p
	}
}

func WithBlinnPhongFlatShading(enable bool) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.flatShading = enable
	}
}

func WithBlinnPhongAmbientOcclusion(enable bool) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.ambientOcclusion = enable
	}
}

func WithBlinnPhongShadow(enable bool) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.receiveShadow = enable
	}
}

func NewBlinnPhong(opts ...BlinnPhongMaterialOption) Material {
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
	matModel := uniforms["matModel"].(math.Matrix)
	matView := uniforms["matView"].(math.Matrix)
	matProj := uniforms["matProj"].(math.Matrix)
	matVP := uniforms["matVP"].(math.Matrix)
	matNormal := uniforms["matNormal"].(math.Matrix)

	pos := v.Pos.Apply(matModel).Apply(matView).Apply(matProj).Apply(matVP).Pos()
	return primitive.Vertex{
		Pos: pos,
		Col: v.Col,
		UV:  v.UV,
		Nor: v.Nor.Apply(matNormal),
	}
}

func (m *BlinnPhongMaterial) FragmentShader(col color.RGBA, x, n, fN, c math.Vector, ls []light.Source, es []light.Environment) color.RGBA {
	LaR := 0.0
	LaG := 0.0
	LaB := 0.0

	for _, e := range es {
		LaR += e.Intensity() * float64(col.R)
		LaG += e.Intensity() * float64(col.G)
		LaB += e.Intensity() * float64(col.B)
	}

	LdR := 0.0
	LdG := 0.0
	LdB := 0.0

	LsR := 0.0
	LsG := 0.0
	LsB := 0.0

	if m.flatShading {
		n = fN
	}

	for _, l := range ls {
		L := l.Position().Sub(x).Unit()
		V := c.Sub(x).Unit()
		H := L.Add(V).Unit()
		Ld := math.Clamp(n.Dot(L), 0, 1)
		Ls := math.Pow(math.Clamp(n.Dot(H), 0, 1), m.shininess)
		I := l.Intensity() / l.Position().Sub(x).Len()

		LdR += Ld * float64(col.R) * I
		LdG += Ld * float64(col.G) * I
		LdB += Ld * float64(col.B) * I

		LsR += Ls * float64(l.Color().R) * I
		LsG += Ls * float64(l.Color().G) * I
		LsB += Ls * float64(l.Color().B) * I
	}

	// The Blinn-Phong Reflection Model
	r := LaR + m.kDiff*LdR + m.kSpec*LsR
	g := LaG + m.kDiff*LdG + m.kSpec*LsG
	b := LaB + m.kDiff*LdB + m.kSpec*LsB

	return color.RGBA{
		uint8(math.Clamp(r, 0, 0xff)),
		uint8(math.Clamp(g, 0, 0xff)),
		uint8(math.Clamp(b, 0, 0xff)),
		uint8(math.Clamp(float64(col.A), 0, 0xff))}
}

func (m *BlinnPhongMaterial) ReceiveShadow() bool {
	return m.receiveShadow
}

func (m *BlinnPhongMaterial) AmbientOcclusion() bool {
	return m.ambientOcclusion
}
