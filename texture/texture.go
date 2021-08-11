// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

import (
	"image"

	"poly.red/color"
	"poly.red/math"
)

// texture implements Texture and represents a power-of-two 2D texture.
// The power-of-two means that the texture width and height must be a
// power of two. e.g. 1024x1024.
type texture struct {
	sampleOption

	format    Format
	encode    Encode
	useMipmap bool
	mipmap    []*image.RGBA
	data      *image.RGBA
	debug     bool
}

func newTexture(data *image.RGBA) Texture {

	return &texture{data: data}
}

// Size returns the size of the texture.
func (t *texture) Size() int {
	return t.data.Bounds().Dx()
}

// UseMipmap checks if the texture activates mipmap.
func (t *texture) UseMipmap() bool {
	return t.useMipmap
}

// Sample fetches the color of at pixel (u, v). This function is a naive
// mipmap implementation that does magnification and minification.
func (t *texture) Sample(sampler Sampler, coord math.Vec2) color.RGBA {
	iu, u := math.Modf(coord.X)
	if iu != 0 && u == 0 {
		u = 1
	}
	if u < 0 {
		u = 1 - u
	}

	iv, v := math.Modf(coord.Y)
	if iv != 0 && v == 0 {
		v = 1
	}
	if v < 0 {
		v = 1 - v
	}

	if !t.useMipmap {
		return t.queryL0(u, v)
	}

	// FIXME: what to do with sampler??
	lod := 0.0

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
	if math.ApproxEq(p, 0, math.Epsilon) {
		return t.queryBilinear(h, u, v)
	}
	return t.queryTrilinear(h, l, p, u, v)
}

func (t *texture) queryL0(u, v float64) color.RGBA {
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

func (t *texture) queryTrilinear(h, l int, p, u, v float64) color.RGBA {
	L1 := t.queryBilinear(h, u, v)
	L2 := t.queryBilinear(l, u, v)
	return math.LerpC(L1, L2, p)
}

func (t *texture) queryBilinear(lod int, u, v float64) color.RGBA {
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
