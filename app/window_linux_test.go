// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"image"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"

	"poly.red/gpu"
)

// requireOrSkip turns a skip into a hard failure when POLYRED_REQUIRE_WINDOW is
// set. CI runs the windowed test in an environment where the display and the GL
// runtime are guaranteed present (Xvfb + Mesa), so a skip there means the very
// thing the test exists to prove silently did not run. On a bare dev box the env
// var is unset and the test skips cleanly.
func requireOrSkip(t *testing.T, format string, args ...any) {
	t.Helper()
	if os.Getenv("POLYRED_REQUIRE_WINDOW") != "" {
		t.Fatalf("POLYRED_REQUIRE_WINDOW set but the windowed path is unavailable: "+format, args...)
	}
	t.Skipf(format, args...)
}

// solidRGBA returns a tightly-packed w*h RGBA image filled with c.
func solidRGBA(w, h int, c [4]byte) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for i := 0; i < len(img.Pix); i += 4 {
		img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = c[0], c[1], c[2], c[3]
	}
	return img
}

// TestX11WindowedPresent drives the cgo-free X11 + GPU-Device windowed present
// path end to end: open an X display, create+map a window, open the GL device,
// bind an on-screen Surface to the window, and present several frames across a
// resize, reading the presented pixels back each time. It is the runtime proof of
// the present path AND the thread/context-ownership model: all GL/EGL runs on the
// backend's single locked thread while the app drives present from another, so a
// per-thread current-context bug would deadlock or render wrong here (a single
// clear-and-readback would miss it -- hence multiple frames + a resize).
//
// It runs under Xvfb + Mesa llvmpipe in CI (the gl-probe x11-windowed-present job)
// with POLYRED_REQUIRE_WINDOW=1 so a skip is a failure there; on a bare dev box it
// skips cleanly.
func TestX11WindowedPresent(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		requireOrSkip(t, "no X display (set DISPLAY / run under Xvfb)")
	}
	// X11 is thread-bound; pin this goroutine like run() does. (The GL backend
	// owns its own locked thread; present marshals onto it.)
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := loadX11(); err != nil {
		requireOrSkip(t, "libX11 unavailable: %v", err)
	}
	d, _, _ := purego.SyscallN(_XOpenDisplay, 0)
	if d == 0 {
		requireOrSkip(t, "XOpenDisplay returned NULL (no reachable X server)")
	}
	display := uintptr(d)

	const w, h = 64, 48
	swa := x11SetWindowAttributes{
		eventMask:        xExposureMask | xStructureNotifyMask,
		backgroundPixmap: xNone,
		overrideRedirect: xFalse,
	}
	root, _, _ := purego.SyscallN(_XDefaultRootWindow, display)
	oswin, _, _ := purego.SyscallN(_XCreateWindow,
		display, root, 0, 0, uintptr(w), uintptr(h),
		0, xCopyFromParent, xInputOutput, 0,
		xCWEventMask|xCWBackPixmap|xCWOverrideRedirect,
		uintptr(unsafe.Pointer(&swa)))
	runtime.KeepAlive(&swa)
	if oswin == 0 {
		t.Fatal("XCreateWindow returned 0 (window creation failed)")
	}
	window := uint64(oswin)
	purego.SyscallN(_XMapWindow, display, uintptr(window))
	defer func() {
		purego.SyscallN(_XDestroyWindow, display, uintptr(window))
		purego.SyscallN(_XCloseDisplay, display)
	}()

	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		requireOrSkip(t, "no GL device (libEGL/libGLESv2/driver missing): %v", err)
	}
	defer dev.Close()

	surf, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Display: display,
		Window:  uintptr(window),
		Width:   w,
		Height:  h,
		Format:  gpu.RGBA8Unorm,
	})
	if err != nil {
		// eglCreateWindowSurface failing here is the classic X11-visual / EGL-config
		// mismatch -- the thing this test exists to catch once we have a device.
		t.Fatalf("CreateWindowSurface failed (X11 visual / EGL config mismatch): %v", err)
	}
	defer surf.Release()

	red := [4]byte{255, 0, 0, 255}

	// presentAndCheck presents a solid-red frame of size sw x sh and asserts the
	// presented pixels read back red. Driving this several times and across a
	// resize exercises the present loop and the resize realloc on the backend
	// thread, not just a one-shot.
	presentAndCheck := func(sw, sh int) {
		img := solidRGBA(sw, sh, red)
		if err := surf.PresentImage(img); err != nil {
			t.Fatalf("PresentImage(%dx%d) failed: %v", sw, sh, err)
		}
		pix := surf.PresentedPixels()
		if len(pix) != sw*sh*4 {
			t.Fatalf("PresentedPixels len=%d, want %d", len(pix), sw*sh*4)
		}
		// Check the center pixel (and a corner) to catch all-black / channel-swap.
		off := ((sh/2)*sw + sw/2) * 4
		got := [4]byte{pix[off], pix[off+1], pix[off+2], pix[off+3]}
		for i := range red {
			if diff := int(got[i]) - int(red[i]); diff < -2 || diff > 2 {
				t.Fatalf("presented center pixel=%v, want ~%v (gl present/blit marshaling)", got, red)
			}
		}
	}

	// Several frames at the original size.
	for range 4 {
		presentAndCheck(w, h)
	}

	// Resize the swapchain and present several more frames. surf.Resize reallocates
	// the upload/blit texture on the backend thread; if the thread/context model is
	// wrong this is where it shows up.
	const w2, h2 = 48, 32
	if err := surf.Resize(w2, h2); err != nil {
		t.Fatalf("surface Resize failed: %v", err)
	}
	for range 4 {
		presentAndCheck(w2, h2)
	}
}
