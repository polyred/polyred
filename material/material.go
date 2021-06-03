package material

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

type Material interface {
	Texture() *Texture
	Wireframe() color.RGBA
	Shader(col color.RGBA, x, n, l, camera math.Vector) color.RGBA
}

type BlinnPhongMaterial struct {
	tex       *Texture
	wireframe color.RGBA

	kDiff     float64
	kSpec     float64
	kAmb      float64
	shininess float64
}

func NewBlinnPhongMaterial(t *Texture, w color.RGBA, Kdiff, Kspec, Kamb, shininess float64) Material {
	return &BlinnPhongMaterial{
		tex:       t,
		wireframe: w,
		kDiff:     Kdiff,
		kSpec:     Kspec,
		kAmb:      Kamb,
		shininess: shininess,
	}
}

func (m *BlinnPhongMaterial) Texture() *Texture {
	return m.tex
}

func (m *BlinnPhongMaterial) Wireframe() color.RGBA {
	return m.wireframe
}

func (m *BlinnPhongMaterial) Shader(col color.RGBA, x, n, l, c math.Vector) color.RGBA {
	L := l.Sub(x).Unit()
	V := c.Sub(x).Unit()
	H := L.Add(V).Unit()
	p := m.shininess
	La := math.Clamp(m.kAmb, 0, 255)
	Ld := math.Clamp(m.kDiff*n.Dot(L), 0, 255)
	Ls := math.Clamp(m.kSpec*math.Pow(n.Dot(H), p), 0, 255)
	shade := La + Ld + Ls
	r := uint8(math.Clamp(shade*float64(col.R), 0, 255))
	g := uint8(math.Clamp(shade*float64(col.G), 0, 255))
	b := uint8(math.Clamp(shade*float64(col.B), 0, 255))
	return color.RGBA{r, g, b, col.A}
}
