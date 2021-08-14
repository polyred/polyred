// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package buffer

import (
	"image"
	"sync"

	"poly.red/geometry/primitive"
)

// Fragment is a collection regarding the relevant geometry information of a fragment.
type Fragment struct {
	Ok bool // true if ok for access or false otherwise
	primitive.Fragment
}

// Buffer is a rendering buffer that supports concurrent-safe
// depth testing and pixel operation.
type Buffer struct {
	lock      []sync.Mutex
	fragments []Fragment
	stride    int
	rect      image.Rectangle

	format PixelFormat // define the alignment of color values

	// TODO: using []uint8 to better support pixel format
	depth []uint8
	color []uint8
}

type PixelFormat int

const (
	PixelFormatRGBA PixelFormat = iota
	PixelFormatBGRA
)

// Opt is a buffer option
type Opt func(b *Buffer)

// Format returns a pixel format option
func Format(format PixelFormat) Opt {
	return func(b *Buffer) {
		b.format = format
	}
}

func NewBuffer(r image.Rectangle, opts ...Opt) *Buffer {
	buf := &Buffer{
		lock:      make([]sync.Mutex, r.Dx()*r.Dy()),
		depth:     make([]uint8, 4*r.Dx()*r.Dy()),
		color:     make([]uint8, 4*r.Dx()*r.Dy()),
		fragments: make([]Fragment, r.Dx()*r.Dy()),
		stride:    r.Dx(),
		rect:      r,
	}

	for _, opt := range opts {
		opt(buf)
	}
	return buf
}

func (b *Buffer) Clear() {
	// Clear using zero values.
	// This loop involves compiler optimization, see:
	// https://golang.org/issue/5373

	for i := range b.fragments {
		b.fragments[i] = Fragment{}
	}
	for i := range b.depth {
		b.depth[i] = 0
	}
	for i := range b.color {
		b.color[i] = 0
	}
}

func (b *Buffer) Image() *image.RGBA {
	return &image.RGBA{
		Stride: 4 * b.stride,
		Rect:   b.rect,
		Pix:    b.color,
	}
}

func (b *Buffer) Depth() *image.RGBA {
	return &image.RGBA{
		Stride: 4 * b.stride,
		Rect:   b.rect,
		Pix:    b.depth,
	}
}

func (b *Buffer) Bounds() image.Rectangle { return b.rect }

func (b *Buffer) fragmentOffset(x, y int) int {
	return (y-b.rect.Min.Y)*b.stride + (x - b.rect.Min.X)
}

func (b *Buffer) pixelOffset(x, y int) int {
	return (y-b.rect.Min.Y)*b.stride*4 + (x-b.rect.Min.X)*4
}

func (b *Buffer) In(x, y int) bool {
	return image.Point{x, y}.In(b.rect)
}

func (b *Buffer) At(x, y int) Fragment {
	if !(image.Point{x, b.rect.Max.Y - y}.In(b.rect)) {
		return Fragment{}
	}
	i := b.fragmentOffset(x, b.rect.Max.Y-y)

	b.lock[i].Lock()
	info := b.fragments[i]
	b.lock[i].Unlock()
	return info
}

func (b *Buffer) Set(x, y int, info Fragment) {
	if !(image.Point{x, b.rect.Max.Y - y}.In(b.rect)) {
		return
	}
	i := b.fragmentOffset(x, b.rect.Max.Y-y)

	b.lock[i].Lock()
	defer b.lock[i].Unlock()

	if b.fragments[i].Ok && info.Depth <= b.fragments[i].Depth {
		return
	}

	j := b.pixelOffset(x, b.rect.Max.Y-y)

	// Write color and depth information to the two dedicated color and
	// depth buffers.
	d := b.depth[j : j+4 : j+4] // Small cap improves performance, see https://golang.org/issue/27857
	c := b.color[j : j+4 : j+4]

	switch b.format {
	case PixelFormatBGRA:
		d[2] = uint8(info.Depth * 0xff)
		d[1] = uint8(info.Depth * 0xff)
		d[0] = uint8(info.Depth * 0xff)
		d[3] = 0xff

		c[2] = info.Col.R
		c[1] = info.Col.G
		c[0] = info.Col.B
		c[3] = info.Col.A
	default: // PixelFormatRGBA:
		d[0] = uint8(info.Depth * 0xff)
		d[1] = uint8(info.Depth * 0xff)
		d[2] = uint8(info.Depth * 0xff)
		d[3] = 0xff

		c[0] = info.Col.R
		c[1] = info.Col.G
		c[2] = info.Col.B
		c[3] = info.Col.A
	}

	b.fragments[i] = info
}

// DepthTest conducts the depth test.
func (b *Buffer) DepthTest(x, y int, depth float64) bool {
	if !(image.Point{x, b.rect.Max.Y - y}.In(b.rect)) {
		return false
	}
	i := b.fragmentOffset(x, b.rect.Max.Y-y)

	b.lock[i].Lock()
	defer b.lock[i].Unlock()
	// If the fragments is not ok to use, or the depth greater than the
	// existing depth value, pass the test.
	return (!b.fragments[i].Ok) || depth > b.fragments[i].Depth
}
