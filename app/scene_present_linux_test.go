// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"os"
	"runtime"
	"testing"

	"github.com/ebitengine/purego"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/gpu"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

// TestScenePresentReadback is the full scene -> present -> readback e2e: it
// renders a real scene (the bunny) to an *image.RGBA with the CPU renderer, then
// drives the exact on-screen path the app uses -- create an X11 window, open the
// GL device on it, bind a Surface, PresentImage the rendered frame, and read the
// presented pixels back -- and asserts the window shows what was rendered. The
// two halves were each covered before (scene->image in internal/examples;
// image->window in TestX11WindowedPresent); this stitches them into one
// continuous check that a rendered scene actually reaches the screen intact.
//
// The present is a same-size NEAREST blit of a linear RGBA8 texture to a linear
// (non-sRGB) window framebuffer, so the readback should reproduce the rendered
// image almost exactly; the gate is tight. Runs in the dedicated Xvfb + Mesa job
// (POLYRED_REQUIRE_WINDOW); a no-op elsewhere.
func TestScenePresentReadback(t *testing.T) {
	if os.Getenv("POLYRED_REQUIRE_WINDOW") == "" {
		t.Skip("scene present e2e runs only in the dedicated Xvfb+Mesa job (POLYRED_REQUIRE_WINDOW)")
	}
	if os.Getenv("DISPLAY") == "" {
		requireOrSkip(t, "no X display (set DISPLAY / run under Xvfb)")
	}
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	const w, h = 96, 96

	// 1. Render a real scene to an image on the CPU (independent of the present
	// device, and deterministic with a single worker).
	s, c := presentTestScene(w, h)
	img := render.NewRenderer(
		render.Scene(s), render.Camera(c), render.Size(w, h),
		render.Background(color.RGBA{0, 127, 255, 255}),
		render.Workers(1), render.CPU(),
	).Render()
	if len(img.Pix) != w*h*4 {
		t.Fatalf("rendered image is %d bytes, want %d", len(img.Pix), w*h*4)
	}
	// Guard against a blank render trivially passing the readback match: the lit
	// bunny on a blue background must have real spatial variation.
	if varied := countVaried(img.Pix, img.Pix[0:4]); varied < len(img.Pix)/4/20 {
		t.Fatalf("rendered scene looks blank (%d/%d pixels differ from the corner) -- render may have failed", varied, len(img.Pix)/4)
	}

	// 2. Create the X11 window + GL device + on-screen surface, exactly as run().
	if err := loadX11(); err != nil {
		requireOrSkip(t, "libX11 unavailable: %v", err)
	}
	d, _, _ := purego.SyscallN(_XOpenDisplay, 0)
	if d == 0 {
		requireOrSkip(t, "XOpenDisplay returned NULL")
	}
	display := uintptr(d)
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL), gpu.WithNativeDisplay(display))
	if err != nil {
		requireOrSkip(t, "no GL device: %v", err)
	}
	defer dev.Close()
	window, err := createX11Window(display, dev.WindowVisualID(), w, h)
	if err != nil {
		t.Fatalf("createX11Window: %v", err)
	}
	defer func() {
		purego.SyscallN(_XDestroyWindow, display, uintptr(window))
		purego.SyscallN(_XCloseDisplay, display)
	}()
	surf, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Display: display, Window: uintptr(window), Width: w, Height: h, Format: gpu.RGBA8Unorm,
	})
	if err != nil {
		requireOrSkip(t, "CreateWindowSurface failed: %v", err)
	}
	defer surf.Release()

	// 3. Present the rendered scene and read the presented pixels back.
	if err := surf.PresentImage(img); err != nil {
		t.Fatalf("PresentImage(rendered scene): %v", err)
	}
	got := surf.PresentedPixels()
	if len(got) != len(img.Pix) {
		t.Fatalf("PresentedPixels len=%d, want %d", len(got), len(img.Pix))
	}

	// 4. The window must show what was rendered. Same-size linear blit, so expect a
	// near-exact reproduction.
	nBig := 0
	for i := range img.Pix {
		diff := int(got[i]) - int(img.Pix[i])
		if diff < 0 {
			diff = -diff
		}
		if diff > 2 {
			nBig++
		}
	}
	if frac := float64(nBig) / float64(len(img.Pix)); frac > 0.01 {
		t.Fatalf("presented scene differs from the rendered scene: %.2f%% of channels off by >2 (want <1%%)", frac*100)
	}
}

// presentTestScene builds the standard lit-bunny test scene used across the
// renderer tests, at the given size.
func presentTestScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(
		light.NewPoint(
			light.Intensity(5),
			light.Color(color.RGBA{0, 0, 0, 255}),
			light.Position(math.NewVec3[float32](-2, 2.5, 6)),
		),
		light.NewAmbient(light.Intensity(0.5)),
	)
	m := model.MustLoad("../internal/testdata/bunny.obj")
	m.Rotate(math.NewVec3[float32](0, 1, 0), -math.Pi/6)
	m.Scale(4, 4, 4)
	m.Translate(0.1, 0, -0.2)
	s.Add(m)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 1.5, 1)),
		camera.LookAt(
			math.NewVec3[float32](0, 0, -0.5),
			math.NewVec3[float32](0, 1, 0),
		),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 3),
	)
}

// countVaried returns how many RGBA pixels in pix differ from ref by more than a
// small threshold in any channel.
func countVaried(pix []byte, ref []byte) int {
	n := 0
	for i := 0; i+4 <= len(pix); i += 4 {
		for c := range 4 {
			d := int(pix[i+c]) - int(ref[c])
			if d < 0 {
				d = -d
			}
			if d > 16 {
				n++
				break
			}
		}
	}
	return n
}
