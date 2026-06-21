// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package ca_test

import (
	"testing"
	"unsafe"

	"github.com/ebitengine/purego/objc"

	"poly.red/gpu/ctx/ca"
	"poly.red/gpu/mtl"
)

// TestMetalLayerOffscreen exercises the cgo-free CAMetalLayer operations against
// an off-screen layer (no window): the setters must not crash and the pixel
// format round-trips. On-screen present (a drawable from NextDrawable attached to
// a view) needs a display and is verified by running a windowed app, not here.
func TestMetalLayerOffscreen(t *testing.T) {
	dev, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		t.Skipf("no Metal device: %v", err)
	}
	cls := objc.GetClass("CAMetalLayer")
	if cls == 0 {
		t.Skip("CAMetalLayer class unavailable")
	}
	layerID := objc.ID(cls).Send(objc.RegisterName("alloc")).Send(objc.RegisterName("init"))
	if layerID == 0 {
		t.Fatal("could not create a CAMetalLayer")
	}
	defer layerID.Send(objc.RegisterName("release"))

	ml := ca.NewMetalLayer(unsafe.Pointer(layerID))
	ml.SetDevice(dev)
	ml.SetPixelFormat(mtl.PixelFormatBGRA8UNorm)
	ml.SetDrawableSize(64, 64)
	ml.SetMaximumDrawableCount(2)
	ml.SetDisplaySyncEnabled(false)

	if pf := ml.PixelFormat(); pf != mtl.PixelFormatBGRA8UNorm {
		t.Errorf("PixelFormat round-trip = %v, want %v", pf, mtl.PixelFormatBGRA8UNorm)
	}

	// An off-screen layer (not in a view tree) typically has no drawable; either
	// outcome is fine here, we just require the call not to crash.
	if d, err := ml.NextDrawable(); err == nil {
		_ = d.Texture()
		_ = d.Drawable()
	}
}
