// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

// Package ca provides cgo-free access to CAMetalLayer (the present surface for
// windowed Metal), through purego/objc -- the same approach as poly.red/gpu/mtl.
// The layer itself is obtained from a native view (still created by the platform
// window layer); this package only messages it.
// https://developer.apple.com/documentation/quartzcore/cametallayer
package ca

import (
	"errors"
	"unsafe"

	"github.com/ebitengine/purego/objc"

	"poly.red/gpu/mtl"
)

var (
	selSetDevice             = objc.RegisterName("setDevice:")
	selSetPixelFormat        = objc.RegisterName("setPixelFormat:")
	selPixelFormat           = objc.RegisterName("pixelFormat")
	selSetMaxDrawableCount   = objc.RegisterName("setMaximumDrawableCount:")
	selSetDisplaySyncEnabled = objc.RegisterName("setDisplaySyncEnabled:")
	selSetDrawableSize       = objc.RegisterName("setDrawableSize:")
	selNextDrawable          = objc.RegisterName("nextDrawable")
	selDrawableTexture       = objc.RegisterName("texture")
)

// cgSize matches CGSize (two CGFloat == two float64 on 64-bit), passed by value.
type cgSize struct{ width, height float64 }

// MetalLayer is a Core Animation Metal layer: a pool of presentable drawables.
type MetalLayer struct {
	layer objc.ID
}

// NewMetalLayer wraps a CAMetalLayer pointer obtained from a native view.
func NewMetalLayer(layer unsafe.Pointer) MetalLayer {
	return MetalLayer{layer: objc.ID(uintptr(layer))}
}

// PixelFormat returns the layer's drawable pixel format.
func (ml MetalLayer) PixelFormat() mtl.PixelFormat {
	return mtl.PixelFormat(uint8(objc.Send[uint64](ml.layer, selPixelFormat)))
}

// SetDevice sets the Metal device responsible for the layer's drawables.
func (ml MetalLayer) SetDevice(device mtl.Device) {
	ml.layer.Send(selSetDevice, objc.ID(uintptr(device.Device())))
}

// SetPixelFormat sets the layer's drawable pixel format (BGRA8UNorm and a few
// others are valid for a Metal layer).
func (ml MetalLayer) SetPixelFormat(pf mtl.PixelFormat) {
	ml.layer.Send(selSetPixelFormat, uint64(pf))
}

// SetMaximumDrawableCount sets the drawable pool size (2 or 3).
func (ml MetalLayer) SetMaximumDrawableCount(count int) {
	ml.layer.Send(selSetMaxDrawableCount, uint64(count))
}

// SetDisplaySyncEnabled controls vsync for the layer's drawables.
func (ml MetalLayer) SetDisplaySyncEnabled(enabled bool) {
	var b uint64
	if enabled {
		b = 1
	}
	ml.layer.Send(selSetDisplaySyncEnabled, b)
}

// SetDrawableSize sets the size, in pixels, of the layer's drawables.
func (ml MetalLayer) SetDrawableSize(width, height int) {
	ml.layer.Send(selSetDrawableSize, cgSize{float64(width), float64(height)})
}

// NextDrawable returns the next presentable drawable, or an error if none is
// available (for example an off-screen layer not attached to a view).
func (ml MetalLayer) NextDrawable() (caMetalDrawable, error) {
	md := ml.layer.Send(selNextDrawable)
	if md == 0 {
		return caMetalDrawable{}, errors.New("ca: nextDrawable returned nil")
	}
	return caMetalDrawable{md}, nil
}

// caMetalDrawable is a displayable resource Metal can render into.
type caMetalDrawable struct {
	drawable objc.ID
}

// Drawable implements the Drawable interface (the raw CAMetalDrawable pointer).
func (md caMetalDrawable) Drawable() unsafe.Pointer {
	return unsafe.Pointer(md.drawable)
}

// Texture returns the drawable's backing Metal texture.
func (md caMetalDrawable) Texture() mtl.Texture {
	return mtl.NewTexture(unsafe.Pointer(md.drawable.Send(selDrawableTexture)))
}
