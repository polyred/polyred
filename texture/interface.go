// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

import (
	"fmt"
	"image"

	"poly.red/color"
	"poly.red/math"
	"poly.red/utils"
)

type SampleOption func(opt *sampleOption)

type sampleOption struct {
	wrapS     Wrap
	wrapT     Wrap
	magFilter Filter
	minFilter Filter
}

func SampleWrap(s, t Wrap) SampleOption {
	return func(opt *sampleOption) {
		opt.wrapS = s
		opt.wrapT = t
	}
}

func SampleFilter(mag, min Filter) SampleOption {
	return func(opt *sampleOption) {
		opt.magFilter = mag
		opt.minFilter = min
	}
}

type Texture interface {
	Sample(coord math.Vec2, opts ...SampleOption) color.RGBA
}

func New(opts ...TextureOption) Texture {
	t := &texture{
		useMipmap: true,
		data:      defaultTexture,
		mipmap:    []*image.RGBA{},
	}
	for _, opt := range opts {
		opt(t)
	}

	dx := t.data.Bounds().Dx()
	dy := t.data.Bounds().Dy()
	if dx == 1 && dy == 1 {
		t.mipmap = []*image.RGBA{t.data}
		return t
	}

	L := int(math.Log2(math.Max(float64(dx), float64(dy)))) + 1
	t.mipmap = make([]*image.RGBA, L)
	t.mipmap[0] = t.data

	for i := 1; i < L; i++ {
		width := dx / int(math.Pow(2, float64(i)))
		height := dy / int(math.Pow(2, float64(i)))
		t.mipmap[i] = utils.Resize(width, height, t.data)
		if t.debug {
			utils.SaveImage(t.mipmap[i], fmt.Sprintf("%d.png", i))
		}
	}
	return t
}

var defaultTexture = &image.RGBA{
	Pix:    []uint8{255, 255, 255, 255},
	Stride: 4,
	Rect:   image.Rect(0, 0, 1, 1),
}
