// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// +build darwin

// Package coreanim provides access to Apple's Core Animation API.
// https://developer.apple.com/documentation/quartzcore
// https://developer.apple.com/documentation/appkit
package mtl

import (
	"errors"
	"unsafe"
)

/*
#cgo LDFLAGS: -framework QuartzCore -framework Foundation
#include <stdbool.h>
#include "coreanim.h"
*/
import "C"

// Window is a window that an app displays on the screen.
// https://developer.apple.com/documentation/appkit/nswindow.
type Window struct {
	window unsafe.Pointer
}

// NewWindow returns a Window that wraps an existing NSWindow * pointer.
func NewWindow(window unsafe.Pointer) Window {
	return Window{window}
}

// ContentView returns the window's content view, the highest accessible View
// in the window's view hierarchy.
// https://developer.apple.com/documentation/appkit/nswindow/1419160-contentview.
func (w Window) ContentView() View {
	return View{C.Window_ContentView(w.window)}
}

// View is the infrastructure for drawing, printing, and handling events in an app.
// https://developer.apple.com/documentation/appkit/nsview.
type View struct {
	view unsafe.Pointer
}

// SetLayer sets v.layer to l.
// https://developer.apple.com/documentation/appkit/nsview/1483298-layer.
func (v View) SetLayer(l Layer) {
	C.View_SetLayer(v.view, l.Layer())
}

// SetWantsLayer sets v.wantsLayer to wantsLayer.
// https://developer.apple.com/documentation/appkit/nsview/1483695-wantslayer.
func (v View) SetWantsLayer(wantsLayer bool) {
	C.View_SetWantsLayer(v.view, C.bool(wantsLayer))
}

// Layer is an object that manages image-based content and
// allows you to perform animations on that content.
// https://developer.apple.com/documentation/quartzcore/calayer.
type Layer interface {
	// Layer returns the underlying CALayer * pointer.
	Layer() unsafe.Pointer
}

// MetalLayer is a Core Animation Metal layer, a layer that manages a pool of Metal drawables.
// https://developer.apple.com/documentation/quartzcore/cametallayer.
type MetalLayer struct {
	metalLayer unsafe.Pointer
}

// MakeMetalLayer creates a new Core Animation Metal layer.
// https://developer.apple.com/documentation/quartzcore/cametallayer.
func MakeMetalLayer() MetalLayer {
	return MetalLayer{C.MakeMetalLayer()}
}

// Layer implements the Layer interface.
func (ml MetalLayer) Layer() unsafe.Pointer { return ml.metalLayer }

// PixelFormat returns the pixel format of textures for rendering layer content.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478155-pixelformat.
func (ml MetalLayer) PixelFormat() PixelFormat {
	return PixelFormat(C.MetalLayer_PixelFormat(ml.metalLayer))
}

// SetDevice sets the Metal device responsible for the layer's drawable resources.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478163-device.
func (ml MetalLayer) SetDevice(device Device) {
	C.MetalLayer_SetDevice(ml.metalLayer, device.Device())
}

// SetPixelFormat controls the pixel format of textures for rendering layer content.
// The pixel format for a Metal layer must be PixelFormatBGRA8UNorm, PixelFormatBGRA8UNormSRGB,
// PixelFormatRGBA16Float, PixelFormatBGRA10XR, or PixelFormatBGRA10XRSRGB.
// SetPixelFormat panics for other values.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478155-pixelformat.
func (ml MetalLayer) SetPixelFormat(pf PixelFormat) {
	e := C.MetalLayer_SetPixelFormat(ml.metalLayer, C.uint16_t(pf))
	if e != nil {
		panic(errors.New(C.GoString(e)))
	}
}

// SetMaximumDrawableCount controls the number of Metal drawables in the resource pool
// managed by Core Animation.
// It can set to 2 or 3 only. SetMaximumDrawableCount panics for other values.
// https://developer.apple.com/documentation/quartzcore/cametallayer/2938720-maximumdrawablecount.
func (ml MetalLayer) SetMaximumDrawableCount(count int) {
	e := C.MetalLayer_SetMaximumDrawableCount(ml.metalLayer, C.uint_t(count))
	if e != nil {
		panic(errors.New(C.GoString(e)))
	}
}

// SetDisplaySyncEnabled controls whether the Metal layer and its drawables
// are synchronized with the display's refresh rate.
// https://developer.apple.com/documentation/quartzcore/cametallayer/2887087-displaysyncenabled.
func (ml MetalLayer) SetDisplaySyncEnabled(enabled bool) {
	C.MetalLayer_SetDisplaySyncEnabled(ml.metalLayer, C.bool(enabled))
}

// SetDrawableSize sets the size, in pixels, of textures for rendering layer content.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478174-drawablesize.
func (ml MetalLayer) SetDrawableSize(width, height int) {
	C.MetalLayer_SetDrawableSize(ml.metalLayer, C.double(width), C.double(height))
}

// NextDrawable returns a Metal drawable.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478172-nextdrawable.
func (ml MetalLayer) NextDrawable() (MetalDrawable, error) {
	md := C.MetalLayer_NextDrawable(ml.metalLayer)
	if md == nil {
		return MetalDrawable{}, errors.New("nextDrawable returned nil")
	}

	return MetalDrawable{md}, nil
}

// MetalDrawable is a displayable resource that can be rendered or written to by Metal.
// https://developer.apple.com/documentation/quartzcore/cametaldrawable.
type MetalDrawable struct {
	metalDrawable unsafe.Pointer
}

// Drawable implements the Drawable interface.
func (md MetalDrawable) Drawable() unsafe.Pointer { return md.metalDrawable }

// Texture returns a Metal texture object representing the drawable object's content.
// https://developer.apple.com/documentation/quartzcore/cametaldrawable/1478159-texture.
func (md MetalDrawable) Texture() Texture {
	return NewTexture(C.MetalDrawable_Texture(md.metalDrawable))
}
