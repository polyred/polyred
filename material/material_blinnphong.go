// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
)

type BlinnPhongMaterial struct {
	tex           *Texture
	kAmb          float64
	kDiff         float64
	kSpec         float64
	shininess     float64
	receiveShadow bool
}

func (m *BlinnPhongMaterial) Texture() *Texture {
	return m.tex
}

type BlinnPhongMaterialOption func(m *BlinnPhongMaterial)

func WithBlinnPhongTexture(tex *Texture) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.tex = tex
	}
}

func WithBlinnPhongFactors(Kamb, Kdiff, Kspec float64) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.kAmb = Kamb
		m.kDiff = Kdiff
		m.kSpec = Kspec
	}
}

func WithBlinnPhongShininess(p float64) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.shininess = p
	}
}

func WithBlinnPhongShadow(enable bool) BlinnPhongMaterialOption {
	return func(m *BlinnPhongMaterial) {
		m.receiveShadow = enable
	}
}

func NewBlinnPhong(opts ...BlinnPhongMaterialOption) Material {
	t := &BlinnPhongMaterial{
		tex:           nil,
		kAmb:          0.5,
		kDiff:         0.5,
		kSpec:         1,
		shininess:     1,
		receiveShadow: false,
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

	pos := v.Pos.Apply(matModel).Apply(matView).Apply(matProj).Apply(matVP)
	return primitive.Vertex{
		Pos: pos.Scale(1/pos.W, 1/pos.W, 1/pos.W, 1/pos.W),
		Col: v.Col,
		UV:  v.UV,
		Nor: v.Nor.Apply(matNormal),
	}
}

func (m *BlinnPhongMaterial) FragmentShader(col color.RGBA, x, n, c math.Vector, ls []light.Light) color.RGBA {
	D := ls[0].Position().Sub(x).Len()
	L := ls[0].Position().Sub(x).Unit()
	V := c.Sub(x).Unit()
	H := L.Add(V).Unit()
	p := m.shininess
	La := m.kAmb
	Ld := m.kDiff * n.Dot(L)
	Ls := m.kSpec * math.Pow(n.Dot(H), p)

	I := ls[0].Itensity() / D

	r := uint8(math.Clamp(math.Round((La+Ld)*float64(col.R)+Ls*float64(ls[0].Color().R)*I), 0, 255))
	g := uint8(math.Clamp(math.Round((La+Ld)*float64(col.G)+Ls*float64(ls[0].Color().G)*I), 0, 255))
	b := uint8(math.Clamp(math.Round((La+Ld)*float64(col.B)+Ls*float64(ls[0].Color().B)*I), 0, 255))
	return color.RGBA{r, g, b, col.A}
}

func (m *BlinnPhongMaterial) ReceiveShadow() bool {
	return m.receiveShadow
}
