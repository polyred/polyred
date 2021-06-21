// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package image

import (
	"fmt"
	"image"
	"image/color"

	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/utils"
)

var defaultTexture = &image.RGBA{
	Pix:    []uint8{255, 255, 255, 255},
	Stride: 4,
	Rect:   image.Rect(0, 0, 1, 1),
}

// Texture represents a power-of-two 2D texture. The power-of-two means
// that the texture width and height must be a power of two. e.g. 1024x1024.
type Texture struct {
	useMipmap bool
	mipmap    []*image.RGBA
	image     *image.RGBA
	debug     bool
}

type TextureOption func(t *Texture)

func WithData(data *image.RGBA) TextureOption {
	return func(t *Texture) {
		if data.Bounds().Dx() < 1 || data.Bounds().Dy() < 1 {
			panic("image width or height is less than 1!")
		}
		t.image = data
	}
}

func WithDebug(enable bool) TextureOption {
	return func(t *Texture) {
		t.debug = enable
	}
}

func WithIsotropicMipMap(enable bool) TextureOption {
	return func(t *Texture) {
		t.useMipmap = enable
	}
}

func NewTexture(opts ...TextureOption) *Texture {
	t := &Texture{
		useMipmap: true,
		image:     defaultTexture,
		mipmap:    []*image.RGBA{},
	}
	for _, opt := range opts {
		opt(t)
	}

	dx := t.image.Bounds().Dx()
	dy := t.image.Bounds().Dy()
	if dx == 1 && dy == 1 {
		t.mipmap = []*image.RGBA{t.image}
		return t
	}

	L := int(math.Log2(math.Max(float64(dx), float64(dy)))) + 1
	t.mipmap = make([]*image.RGBA, L)
	t.mipmap[0] = t.image

	for i := 1; i < L; i++ {
		width := dx / int(math.Pow(2, float64(i)))
		height := dy / int(math.Pow(2, float64(i)))
		t.mipmap[i] = utils.Resize(width, height, t.image)
		if t.debug {
			utils.Save(t.mipmap[i], fmt.Sprintf("%d.png", i))
		}
	}
	return t
}

// Size returns the size of the texture.
func (t *Texture) Size() int {
	return t.image.Bounds().Dx()
}

// UseMipmap checks if the texture activates mipmap.
func (t *Texture) UseMipmap() bool {
	return t.useMipmap
}

// Query fetches the color of at pixel (u, v). This function is a naive
// mipmap implementation that does magnification and minification.
func (t *Texture) Query(lod, u, v float64) color.RGBA {
	iu, u := math.Modf(u)
	if iu != 0 && u == 0 {
		u = 1
	}
	if u < 0 {
		u = 1 - u
	}

	iv, v := math.Modf(v)
	if iv != 0 && v == 0 {
		v = 1
	}
	if v < 0 {
		v = 1 - v
	}

	if !t.useMipmap {
		return t.queryL0(u, v)
	}

	// Make sure LOD is sitting on a valid range before proceed.
	if lod < 0 {
		lod = 0
	} else if lod >= float64(len(t.mipmap)) {
		lod = float64(len(t.mipmap) - 1)
	}

	if lod <= 1 {
		return t.queryBilinear(0, u, v)
	}
	lod -= 1

	// Figure out two different mipmap levels, then compute
	// tri-linear interpolation between the two discrete mipmap levels.
	h := int(math.Floor(lod))
	l := h + 1
	if l >= len(t.mipmap) {
		return t.queryBilinear(h, u, v)
	}

	p := lod - float64(h)
	if math.ApproxEq(p, 0, math.DefaultEpsilon) {
		return t.queryBilinear(h, u, v)
	}
	return t.queryTrilinear(h, l, p, u, v)
}

func (t *Texture) queryL0(u, v float64) color.RGBA {
	tex := t.mipmap[0]
	dx := float64(tex.Bounds().Dx())
	dy := float64(tex.Bounds().Dy())
	if dx == 1 && dy == 1 {
		return tex.At(0, 0).(color.RGBA)
	}

	x := int(math.Floor(u * (dx - 1))) // very coarse approximation.
	y := int(math.Floor(v * (dy - 1))) // very coarse approximation.
	return tex.At(x, y).(color.RGBA)
}

func (t *Texture) queryTrilinear(h, l int, p, u, v float64) color.RGBA {
	L1 := t.queryBilinear(h, u, v)
	L2 := t.queryBilinear(l, u, v)
	return math.LerpC(L1, L2, p)
}

func (t *Texture) queryBilinear(lod int, u, v float64) color.RGBA {
	buf := t.mipmap[lod]
	dx := buf.Bounds().Dx()
	dy := buf.Bounds().Dy()
	if dx == 1 && dy == 1 {
		return buf.At(0, 0).(color.RGBA)
	}
	x := u * (float64(dx) - 1)
	y := v * (float64(dy) - 1)

	x0 := math.Floor(x)
	y0 := math.Floor(y)

	i := int(x0)
	j := int(y0)

	var p1, p2, p3, p4 color.RGBA
	p1 = buf.At(i, j).(color.RGBA)
	if i < dx-1 {
		p2 = buf.At(i+1, j).(color.RGBA)
	} else {
		p2 = buf.At(i, j).(color.RGBA)
	}
	interpo1 := math.LerpC(p1, p2, x-x0)

	if j < dy-1 {
		p3 = buf.At(i, j+1).(color.RGBA)
	} else {
		p3 = buf.At(i, j).(color.RGBA)
	}
	if i < dx-1 && j < dy-1 {
		p4 = buf.At(i+1, j+1).(color.RGBA)
	} else {
		p4 = buf.At(i, j).(color.RGBA)
	}
	interpo2 := math.LerpC(p3, p4, x-x0)
	return math.LerpC(interpo1, interpo2, y-y0)
}
