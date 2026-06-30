// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package render_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/light"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

// demoScene is the scene cmd/simple renders: the Stanford bunny under a point
// light through a perspective camera.
func demoScene() (*scene.Scene, camera.Interface) {
	return scene.NewScene(model.StanfordBunny(), light.NewPoint()), camera.NewPerspective()
}

// nonBlank reports whether an image has real content (not all one value), so a
// smoke test does not pass on a blank or uniform render.
func nonBlank(pix []byte) bool {
	if len(pix) == 0 {
		return false
	}
	first := pix[0]
	for _, b := range pix {
		if b != first {
			return true
		}
	}
	return false
}

// TestRenderSmokeCPU is the basic end-to-end check that the renderer actually
// produces a picture on the CPU path, on every platform. It renders the cmd/simple
// demo scene and asserts a correctly-sized, non-blank image without panicking.
func TestRenderSmokeCPU(t *testing.T) {
	const w, h = 100, 100
	s, c := demoScene()
	img := render.NewRenderer(
		render.Scene(s), render.Camera(c), render.Size(w, h), render.CPU(),
	).Render()
	if img == nil {
		t.Fatal("Render returned nil")
	}
	if got := img.Bounds().Dx(); got != w {
		t.Fatalf("width = %d, want %d", got, w)
	}
	if !nonBlank(img.Pix) {
		t.Fatal("CPU render is blank (the bunny did not draw)")
	}
}

// TestRenderSmokeDefault renders the same scene through the DEFAULT renderer,
// which is GPU-by-default: on macOS it offloads the deferred shading pass to Metal
// (falling back to CPU if no device is available), elsewhere it runs all-CPU. So
// on a macOS CI runner this exercises the GPU path end to end; everywhere it
// guarantees the default path produces a valid picture without error.
func TestRenderSmokeDefault(t *testing.T) {
	const w, h = 100, 100
	s, c := demoScene()
	img := render.NewRenderer(
		render.Scene(s), render.Camera(c), render.Size(w, h),
	).Render()
	if img == nil {
		t.Fatal("Render returned nil")
	}
	if !nonBlank(img.Pix) {
		t.Fatal("default render is blank")
	}
}
