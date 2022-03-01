// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package app

import (
	"errors"
	"unsafe"

	"poly.red/app/internal/mtl"
)

// This file provides access to CAMetalLayer.
// https://developer.apple.com/documentation/quartzcore
// https://developer.apple.com/documentation/appkit

/*
#cgo CFLAGS: -Werror -fmodules -fobjc-arc -x objective-c

#include <AppKit/AppKit.h>
#include <stdbool.h>

typedef unsigned long uint_t;
typedef unsigned short uint16_t;

__attribute__ ((visibility ("hidden"))) CFTypeRef MetalDrawable_Texture(CFTypeRef metalDrawable);
__attribute__ ((visibility ("hidden"))) CFTypeRef MetalLayer_NextDrawable(CFTypeRef metalLayer);
__attribute__ ((visibility ("hidden"))) uint16_t MetalLayer_PixelFormat(CFTypeRef metalLayer);
__attribute__ ((visibility ("hidden"))) void MetalLayer_SetDevice(CFTypeRef metalLayer, CFTypeRef device);
__attribute__ ((visibility ("hidden"))) void MetalLayer_SetDisplaySyncEnabled(CFTypeRef metalLayer, bool displaySyncEnabled);
__attribute__ ((visibility ("hidden"))) void MetalLayer_SetDrawableSize(CFTypeRef metalLayer, double width, double height);
__attribute__ ((visibility ("hidden"))) const char * MetalLayer_SetMaximumDrawableCount(CFTypeRef metalLayer, uint_t maximumDrawableCount);
__attribute__ ((visibility ("hidden"))) const char * MetalLayer_SetPixelFormat(CFTypeRef metalLayer, uint16_t pixelFormat);
*/
import "C"

// caMetalLayer is a Core Animation Metal layer, a layer that manages a pool of Metal drawables.
// https://developer.apple.com/documentation/quartzcore/cametallayer.
type caMetalLayer struct {
	metalLayer C.CFTypeRef
}

// newMetalLayer sets a new Core Animation Metal layer.
// https://developer.apple.com/documentation/quartzcore/cametallayer.
func newMetalLayer(layer C.CFTypeRef) caMetalLayer {
	return caMetalLayer{metalLayer: layer}
}

// PixelFormat returns the pixel format of textures for rendering layer content.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478155-pixelformat.
func (ml caMetalLayer) PixelFormat() mtl.PixelFormat {
	return mtl.PixelFormat(uint8(C.MetalLayer_PixelFormat(ml.metalLayer)))
}

// SetDevice sets the Metal device responsible for the layer's drawable resources.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478163-device.
func (ml caMetalLayer) SetDevice(device mtl.Device) {
	C.MetalLayer_SetDevice(ml.metalLayer, C.CFTypeRef(device.Device()))
}

// SetPixelFormat controls the pixel format of textures for rendering layer content.
// The pixel format for a Metal layer must be PixelFormatBGRA8UNorm, PixelFormatBGRA8UNormSRGB,
// PixelFormatRGBA16Float, PixelFormatBGRA10XR, or PixelFormatBGRA10XRSRGB.
// SetPixelFormat panics for other values.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478155-pixelformat.
func (ml caMetalLayer) SetPixelFormat(pf mtl.PixelFormat) {
	e := C.MetalLayer_SetPixelFormat(ml.metalLayer, C.uint16_t(pf))
	if e != nil {
		panic(C.GoString(e))
	}
}

// SetMaximumDrawableCount controls the number of Metal drawables in the resource pool
// managed by Core Animation.
// It can set to 2 or 3 only. SetMaximumDrawableCount panics for other values.
// https://developer.apple.com/documentation/quartzcore/cametallayer/2938720-maximumdrawablecount.
func (ml caMetalLayer) SetMaximumDrawableCount(count int) {
	e := C.MetalLayer_SetMaximumDrawableCount(ml.metalLayer, C.uint_t(count))
	if e != nil {
		panic(C.GoString(e))
	}
}

// SetDisplaySyncEnabled controls whether the Metal layer and its drawables
// are synchronized with the display's refresh rate.
// https://developer.apple.com/documentation/quartzcore/cametallayer/2887087-displaysyncenabled.
func (ml caMetalLayer) SetDisplaySyncEnabled(enabled bool) {
	C.MetalLayer_SetDisplaySyncEnabled(ml.metalLayer, C.bool(enabled))
}

// SetDrawableSize sets the size, in pixels, of textures for rendering layer content.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478174-drawablesize.
func (ml caMetalLayer) SetDrawableSize(width, height int) {
	C.MetalLayer_SetDrawableSize(ml.metalLayer, C.double(width), C.double(height))
}

// NextDrawable returns a Metal drawable.
// https://developer.apple.com/documentation/quartzcore/cametallayer/1478172-nextdrawable.
func (ml caMetalLayer) NextDrawable() (caMetalDrawable, error) {
	md := C.MetalLayer_NextDrawable(ml.metalLayer)
	if md == 0 {
		return caMetalDrawable{}, errors.New("nextDrawable returned nil")
	}

	return caMetalDrawable{md}, nil
}

// MetalDrawable is a displayable resource that can be rendered or written to by Metal.
// https://developer.apple.com/documentation/quartzcore/cametaldrawable.
type caMetalDrawable struct {
	metalDrawable C.CFTypeRef
}

// Drawable implements the Drawable interface.
func (md caMetalDrawable) Drawable() unsafe.Pointer {
	return unsafe.Pointer(md.metalDrawable)
}

// Texture returns a Metal texture object representing the drawable object's content.
// https://developer.apple.com/documentation/quartzcore/cametaldrawable/1478159-texture.
func (md caMetalDrawable) Texture() mtl.Texture {
	return mtl.NewTexture(unsafe.Pointer(C.MetalDrawable_Texture(C.CFTypeRef(md.metalDrawable))))
}
