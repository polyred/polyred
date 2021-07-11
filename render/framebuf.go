// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"
	"image/color"
	"sync"

	"changkun.de/x/polyred/geometry/primitive"
)

// FragmentInfo is a collection regarding the relevant geometry information of a fragment.
type FragmentInfo struct {
	Ok bool // true if ok for access or false otherwise
	primitive.Fragment
}

// Buffer is a rendering buffer that supports concurrent-safe
// depth testing and pixel operation.
type Buffer struct {
	lock      []sync.Mutex
	fragments []FragmentInfo
	stride    int
	rect      image.Rectangle

	depth *image.RGBA // read only
	color *image.RGBA // read only
}

func NewBuffer(r image.Rectangle) *Buffer {
	buf := &Buffer{
		lock:      make([]sync.Mutex, r.Dx()*r.Dy()),
		depth:     image.NewRGBA(r),
		color:     image.NewRGBA(r),
		fragments: make([]FragmentInfo, r.Dx()*r.Dy()),
		stride:    r.Dx(),
		rect:      r,
	}
	return buf
}

func (b *Buffer) Clear() {
	for i := range b.depth.Pix {
		b.depth.Pix[i] = 0
	}
	for i := range b.color.Pix {
		b.color.Pix[i] = 0
	}
	for i := range b.fragments {
		b.fragments[i] = FragmentInfo{}
	}
}

func (b *Buffer) Image() *image.RGBA {
	return b.color
}

func (b *Buffer) Depth() *image.RGBA {
	return b.depth
}

func (b *Buffer) Bounds() image.Rectangle { return b.rect }

func (b *Buffer) FragmentOffset(x, y int) int {
	return (y-b.rect.Min.Y)*b.stride + (x - b.rect.Min.X)
}

func (b *Buffer) In(x, y int) bool {
	return image.Point{x, y}.In(b.rect)
}

func (b *Buffer) At(x, y int) FragmentInfo {
	if !(image.Point{x, y}.In(b.rect)) {
		return FragmentInfo{}
	}
	i := b.FragmentOffset(x, y)

	b.lock[i].Lock()
	info := b.fragments[i]
	b.lock[i].Unlock()
	return info
}

func (b *Buffer) Set(x, y int, info FragmentInfo) {
	if !(image.Point{x, y}.In(b.rect)) {
		return
	}
	i := b.FragmentOffset(x, y)

	// fast path. depth test fail
	b.lock[i].Lock()
	if b.fragments[i].Ok && info.Depth <= b.fragments[i].Depth {
		b.lock[i].Unlock()
		return
	}
	b.lock[i].Unlock()

	// slow path
	b.lock[i].Lock()
	defer b.lock[i].Unlock()

	if b.fragments[i].Ok && info.Depth <= b.fragments[i].Depth {
		return
	}

	// we also write color and depth information to the two
	// dedicated color and depth buffers.
	b.depth.Set(x, y, color.RGBA{
		uint8(info.Depth * 0xff),
		uint8(info.Depth * 0xff),
		uint8(info.Depth * 0xff), 0xff,
	})
	b.color.Set(x, b.rect.Max.Y-y, info.Col)

	b.fragments[i] = info
}

// DepthTest conducts the depth test.
func (b *Buffer) DepthTest(x, y int, depth float64) bool {
	if !(image.Point{x, y}.In(b.rect)) {
		return false
	}
	i := b.FragmentOffset(x, y)

	b.lock[i].Lock()
	defer b.lock[i].Unlock()
	// If the fragments is not ok to use, or the depth greater than the
	// existing depth value, pass the test.
	return (!b.fragments[i].Ok) || depth > b.fragments[i].Depth
}
