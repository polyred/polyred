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

	"poly.red/gpu/gl"
)

// TestX11WindowedPresent exercises the cgo-free X11 + EGL + GLES windowed-present
// path end to end: open an X display, create and map a window, create an EGL
// window surface bound to it, make a GLES context current, clear to a known
// color, and read the pixels back. It is the only runtime check of the purego
// X11 struct offsets, the EGL window-surface config match, and the gl.Functions
// out-parameter/float marshaling, none of which a cgo-free *build* can prove.
//
// It runs under Xvfb + Mesa llvmpipe in CI (see the gl-probe workflow). It skips
// cleanly when there is no display or no EGL/GLES runtime, so it is a no-op on a
// dev box without that stack rather than a spurious failure.
func TestX11WindowedPresent(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		t.Skip("no X display (set DISPLAY / run under Xvfb)")
	}
	// X11 + GL are thread-bound; pin this goroutine like run()/draw() do.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := loadX11(); err != nil {
		t.Skipf("libX11 unavailable: %v", err)
	}
	d, _, _ := purego.SyscallN(_XOpenDisplay, 0)
	if d == 0 {
		t.Skip("XOpenDisplay returned NULL (no reachable X server)")
	}
	const w, h = 64, 48
	win := &osWindow{
		config:    &config{title: "polyred-test", size: image.Pt(w, h)},
		display:   uintptr(d),
		closed:    make(chan struct{}, 1),
		terminate: make(chan struct{}, 1),
	}

	swa := x11SetWindowAttributes{
		eventMask:        xExposureMask | xStructureNotifyMask,
		backgroundPixmap: xNone,
		overrideRedirect: xFalse,
	}
	root, _, _ := purego.SyscallN(_XDefaultRootWindow, win.display)
	oswin, _, _ := purego.SyscallN(_XCreateWindow,
		win.display, root, 0, 0, uintptr(w), uintptr(h),
		0, xCopyFromParent, xInputOutput, 0,
		xCWEventMask|xCWBackPixmap|xCWOverrideRedirect,
		uintptr(unsafe.Pointer(&swa)))
	runtime.KeepAlive(&swa)
	if oswin == 0 {
		t.Fatal("XCreateWindow returned 0 (window creation failed)")
	}
	win.oswin = uint64(oswin)
	purego.SyscallN(_XMapWindow, win.display, uintptr(win.oswin))
	defer func() {
		purego.SyscallN(_XDestroyWindow, win.display, uintptr(win.oswin))
		purego.SyscallN(_XCloseDisplay, win.display)
	}()

	ctx, err := newX11EGLContext(win)
	if err != nil {
		t.Skipf("no EGL/GLES runtime (libEGL/libGLESv2/driver missing): %v", err)
	}
	win.ctx = ctx
	defer ctx.Release()

	// Refresh creates the EGL window surface (eglCreateWindowSurface). A failure
	// here is the classic X11-visual / EGL-config mismatch, so it is a hard fail
	// once we have a context: it is exactly what this test exists to catch.
	if err := ctx.Refresh(); err != nil {
		t.Fatalf("eglCreateWindowSurface failed (X11 visual / EGL config mismatch): %v", err)
	}
	if err := ctx.Lock(); err != nil {
		t.Fatalf("eglMakeCurrent failed: %v", err)
	}
	defer ctx.Unlock()

	f := ctx.gl
	// String + integer queries exercise the *byte-return and glGetIntegerv
	// out-parameter purego marshaling.
	if v := f.GetString(gl.VERSION); v == "" {
		t.Fatal("glGetString(VERSION) is empty (purego *byte-return marshaling)")
	}
	if m := f.GetInteger(gl.MAX_TEXTURE_SIZE); m <= 0 {
		t.Fatalf("glGetIntegerv(MAX_TEXTURE_SIZE)=%d (out-parameter marshaling)", m)
	}

	// Clear to opaque red and read it back. Red is sRGB-invariant (0 and 1 map to
	// themselves whether or not the surface is sRGB) and channel-specific, so it
	// also catches a channel-swapped readback. ClearColor takes floats, so this
	// is the float-ABI path (purego.RegisterFunc) too.
	f.Viewport(0, 0, w, h)
	f.ClearColor(1, 0, 0, 1)
	f.Clear(gl.COLOR_BUFFER_BIT)
	f.Finish()

	pix := make([]byte, w*h*4)
	f.ReadPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, pix)
	got := [4]byte{pix[0], pix[1], pix[2], pix[3]}
	want := [4]byte{255, 0, 0, 255}
	for i := range want {
		if diff := int(got[i]) - int(want[i]); diff < -2 || diff > 2 {
			t.Fatalf("readback pixel = %v, want ~%v (gl/egl present marshaling)", got, want)
		}
	}

	// Exercise eglSwapBuffers; it must not error on a mapped window surface.
	if err := ctx.Present(); err != nil {
		t.Fatalf("eglSwapBuffers failed: %v", err)
	}
}
