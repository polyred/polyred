// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
)

type Material interface {
	Texture() *Texture
	Wireframe() color.RGBA
	Shader(col color.RGBA, x, n, camera math.Vector, ls []light.Light) color.RGBA
}

type BlinnPhongMaterial struct {
	tex       *Texture
	wireframe color.RGBA

	kAmb      float64
	kDiff     float64
	kSpec     float64
	shininess float64
}

func NewBlinnPhongMaterial(t *Texture, w color.RGBA, Kamb, Kdiff, Kspec, shininess float64) Material {
	return &BlinnPhongMaterial{
		tex:       t,
		wireframe: w,
		kAmb:      Kamb,
		kDiff:     Kdiff,
		kSpec:     Kspec,
		shininess: shininess,
	}
}

func (m *BlinnPhongMaterial) Texture() *Texture {
	return m.tex
}

func (m *BlinnPhongMaterial) Wireframe() color.RGBA {
	return m.wireframe
}

func (m *BlinnPhongMaterial) Shader(col color.RGBA, x, n, c math.Vector, ls []light.Light) color.RGBA {
	D := ls[0].Position().Sub(x).Len()
	L := ls[0].Position().Sub(x).Unit()
	V := c.Sub(x).Unit()
	H := L.Add(V).Unit()
	p := m.shininess
	La := m.kAmb
	Ld := m.kDiff * n.Dot(L)
	Ls := m.kSpec * math.Pow(n.Dot(H), p)

	I := ls[0].Itensity() / D

	r := uint8(math.Clamp((La+Ld)*float64(col.R)+Ls*float64(ls[0].Color().R)*I, 0, 255))
	g := uint8(math.Clamp((La+Ld)*float64(col.G)+Ls*float64(ls[0].Color().G)*I, 0, 255))
	b := uint8(math.Clamp((La+Ld)*float64(col.B)+Ls*float64(ls[0].Color().B)*I, 0, 255))
	return color.RGBA{r, g, b, col.A}
}
