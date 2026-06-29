// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Tests for the on-screen window surface seam. The full pixel-assert test (a
// native window + EGL window surface, asserting the blitted/presented pixels via
// glWindowSurface.readDefault) needs a real display and window-system handle, so
// it is gated to an environment with an X server; the surfaceless CI cannot run
// it. These tests cover the display-independent parts: input validation and the
// backends that do not yet support an on-screen surface (Metal/Vulkan stub it).
package gpu_test

import (
	"errors"
	"testing"

	"poly.red/gpu"
)

func TestCreateWindowSurfaceValidation(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	// Size must be > 0; this is checked before the backend is consulted, so it
	// holds on every driver.
	if _, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Width: 0, Height: 16, Format: gpu.RGBA8Unorm,
	}); err == nil {
		t.Errorf("CreateWindowSurface with Width=0 should error")
	}
	if _, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Width: 16, Height: -1, Format: gpu.RGBA8Unorm,
	}); err == nil {
		t.Errorf("CreateWindowSurface with Height<0 should error")
	}

	// Backends without an on-screen path (Metal, Vulkan) report ErrUnsupported.
	switch dev.Driver() {
	case gpu.DriverMetal, gpu.DriverVulkan:
		_, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
			Width: 16, Height: 16, Format: gpu.RGBA8Unorm,
		})
		if !errors.Is(err, gpu.ErrUnsupported) {
			t.Errorf("CreateWindowSurface on %v = %v, want ErrUnsupported", dev.Driver(), err)
		}
	}
}
