// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package buffer

import (
	"image"

	"poly.red/geometry/primitive"

	"poly.red/internal/spinlock"
)

// Fragment is a collection regarding the relevant geometry information of a fragment.
type Fragment struct {
	Ok bool // true if ok for access or false otherwise
	primitive.Fragment
}

// Buffer is a rendering buffer that supports concurrent-safe
// depth testing and pixel operation.
type Buffer struct {
	lock      []spinlock.SpinLock
	fragments []Fragment
	stride    int
	rect      image.Rectangle

	format PixelFormat // define the alignment of color values
	depth  []uint8
	color  []uint8
}

// PixelFormat represents the internal pixel format of the buffer,
// which determines the order of colors in its internal frame buffer.
type PixelFormat int

// All kinds of pixel format.
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

// NewBuffer returns a rendering buffer. The caller must specify its size.
// By default, it uses RGBA pixel format.
func NewBuffer(r image.Rectangle, opts ...Opt) *Buffer {
	buf := &Buffer{
		lock:      make([]spinlock.SpinLock, r.Dx()*r.Dy()),
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

// Clear clears the entire buffer.
//
// Note that the function is not thread-safe, it is caller's
// responsibility to guarantee that the buffer can be cleared.
func (b *Buffer) Clear() {
	b.ClearFragments()
	b.ClearDepth()
	b.ClearFrameBuf()
}

// ClearFragments clears the buffer's fragments.
//
// Note that the function is not thread-safe, it is caller's
// responsibility to guarantee that the buffer can be cleared.
func (b *Buffer) ClearFragments() {
	// Clear using zero value looping, which involves compiler optimization.
	// See: https://golang.org/issue/5373
	for i := range b.fragments {
		b.fragments[i] = Fragment{}
	}
}

// ClearDepth clears the buffer's internal depth buffer.
//
// Note that the function is not thread-safe, it is caller's
// responsibility to guarantee that the buffer can be cleared.
func (b *Buffer) ClearDepth() {
	// Clear using zero value looping, which involves compiler optimization.
	// See: https://golang.org/issue/5373
	for i := range b.depth {
		b.depth[i] = 0
	}
}

// ClearFrameBuf clears the buffer's frame buffer.
//
// Note that the function is not thread-safe, it is caller's
// responsibility to guarantee that the buffer can be cleared.
func (b *Buffer) ClearFrameBuf() {
	// Clear using zero value looping, which involves compiler optimization.
	// See: https://golang.org/issue/5373
	for i := range b.color {
		b.color[i] = 0
	}
}

func (b *Buffer) Format() PixelFormat {
	return b.format
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

func (b *Buffer) FragmentOffset(x, y int) int {
	return (y-b.rect.Min.Y)*b.stride + (x - b.rect.Min.X)
}

func (b *Buffer) PixelOffset(x, y int) int {
	return (y-b.rect.Min.Y)*b.stride*4 + (x-b.rect.Min.X)*4
}

func (b *Buffer) In(x, y int) bool {
	return image.Point{x, y}.In(b.rect)
}

func (b *Buffer) At(x, y int) Fragment {
	if !(image.Point{x, b.rect.Max.Y - y - 1}.In(b.rect)) {
		return Fragment{}
	}
	i := b.FragmentOffset(x, b.rect.Max.Y-y-1)

	b.lock[i].Lock()
	info := b.fragments[i]
	b.lock[i].Unlock()
	return info
}

func (b *Buffer) Set(x, y int, info Fragment) {
	if !(image.Point{x, b.rect.Max.Y - y - 1}.In(b.rect)) {
		return
	}
	i := b.FragmentOffset(x, b.rect.Max.Y-y-1)

	b.lock[i].Lock()
	defer b.lock[i].Unlock()

	if b.fragments[i].Ok && info.Depth <= b.fragments[i].Depth {
		return
	}

	j := b.PixelOffset(x, b.rect.Max.Y-y-1)

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
	if !(image.Point{x, b.rect.Max.Y - y - 1}.In(b.rect)) {
		return false
	}
	i := b.FragmentOffset(x, b.rect.Max.Y-y-1)

	b.lock[i].Lock()
	defer b.lock[i].Unlock()
	// If the fragments is not ok to use, or the depth greater than the
	// existing depth value, pass the test.
	return (!b.fragments[i].Ok) || depth > b.fragments[i].Depth
}

// UnsafeAt returns a pointer the the underlying fragment without
// bound checks. If the provided pixel coords are invalid, this
// function will result in a panic.
func (b *Buffer) UnsafeAt(x, y int) Fragment {
	i := b.FragmentOffset(x, b.rect.Max.Y-y-1)
	return b.fragments[i]
}

// UnsafeSet sets the given fragment to the underlying frame and
// depth buffer without bound checks. If the provided pixel coords
// are invalid, this function will result in a panic.
func (b *Buffer) UnsafeSet(x, y int, info Fragment) {
	i := b.FragmentOffset(x, b.rect.Max.Y-y-1)
	j := b.PixelOffset(x, b.rect.Max.Y-y-1)

	d := b.depth[j : j+4 : j+4]
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
