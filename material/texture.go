// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"fmt"
	"image"
	"image/color"

	"changkun.de/x/ddd/math"
	"golang.org/x/image/draw"
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
}

type TextureOption func(t *Texture)

func WithImage(img *image.RGBA) TextureOption {
	return func(t *Texture) {
		if img.Bounds().Dx() != img.Bounds().Dy() {
			panic("image width and height is not equal!")
		}
		if img.Bounds().Dx() < 1 {
			panic("image width or height is less than 1!")
		}
		siz := img.Bounds().Dx()
		if (siz&(siz-1)) != 0 || siz < 0 {
			panic("image size is not a power of two!")
		}

		t.image = img
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

	siz := t.image.Bounds().Dx()
	if siz == 1 {
		t.mipmap = []*image.RGBA{t.image}
		return t
	}

	L := int(math.Log2(float64(siz)) + 1)
	t.mipmap = make([]*image.RGBA, L)
	t.mipmap[0] = t.image

	for i := 1; i < L; i++ {
		size := siz / int(math.Pow(2, float64(i)))
		t.mipmap[i] = image.NewRGBA(image.Rect(0, 0, size, size))
		draw.BiLinear.Scale(
			t.mipmap[i], t.mipmap[i].Bounds(),
			t.image, image.Rectangle{
				image.Point{0, 0},
				image.Point{size * 2, size * 2},
			}, draw.Over, nil)
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
func (t *Texture) Query(u, v float64, lod float64) color.RGBA {
	// Early error checking.
	if u < 0 || u > 1 || v < 0 || v > 1 {
		panic(fmt.Sprintf("out of UV query range: %v, %v", u, v))
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
		siz := float64(t.image.Bounds().Dx())
		return t.queryBilinear(0, u*(siz-1), v*(siz-1))
	}
	lod -= 1

	// Figure out two different mipmap levels, then compute
	// tri-linear interpolation between the two discrete mipmap levels.
	h := int(math.Floor(lod))
	l := int(math.Floor(lod)) + 1
	if l >= len(t.mipmap) {
		l = h
	}

	p := lod - float64(h)
	if math.ApproxEq(p, 0, math.DefaultEpsilon) {
		return t.queryBilinear(h, u, v)
	}
	return t.queryTrilinear(h, l, p, u, v)
}

func (t *Texture) queryL0(u, v float64) color.RGBA {
	tex := t.mipmap[0]
	siz := float64(tex.Bounds().Dx())
	if siz == 1 {
		return tex.At(0, 0).(color.RGBA)
	}

	x := int(math.Floor(u * (siz - 1))) // very coarse approximation.
	y := int(math.Floor(v * (siz - 1))) // very coarse approximation.
	return tex.At(x, y).(color.RGBA)
}

func (t *Texture) queryTrilinear(h, l int, p, u, v float64) color.RGBA {
	siz := float64(t.image.Bounds().Dx())
	L1 := t.queryBilinear(
		h,
		(u*(siz-1))/math.Pow(2, float64(h)),
		(v*(siz-1))/math.Pow(2, float64(h)),
	)
	L2 := t.queryBilinear(
		l,
		(u*(siz-1))/math.Pow(2, float64(l)),
		(v*(siz-1))/math.Pow(2, float64(l)),
	)
	return math.LerpC(L1, L2, p)
}

func (t *Texture) queryBilinear(lod int, x, y float64) color.RGBA {
	buf := t.mipmap[lod]
	size := buf.Bounds().Dx()
	if size == 1 {
		return buf.At(0, 0).(color.RGBA)
	}
	x0 := math.Floor(x)
	y0 := math.Floor(y)

	i := int(x0)
	j := int(y0)

	var p1, p2, p3, p4 color.RGBA
	p1 = buf.At(i, j).(color.RGBA)
	if i < size-1 {
		p2 = buf.At(i+1, j).(color.RGBA)
	} else {
		p2 = buf.At(i, j).(color.RGBA)
	}
	interpo1 := math.LerpC(p1, p2, x-x0)

	if j < size-1 {
		p3 = buf.At(i, j+1).(color.RGBA)
	} else {
		p3 = buf.At(i, j).(color.RGBA)
	}
	if i < size-1 && j < size-1 {
		p4 = buf.At(i+1, j+1).(color.RGBA)
	} else {
		p4 = buf.At(i, j).(color.RGBA)
	}
	interpo2 := math.LerpC(p3, p4, x-x0)
	return math.LerpC(interpo1, interpo2, y-y0)
}
