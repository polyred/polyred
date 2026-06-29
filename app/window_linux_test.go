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

// requireOrSkip turns a skip into a hard failure when POLYRED_REQUIRE_WINDOW is
// set. CI runs the windowed test in an environment where the display and the
// EGL/GLES runtime are guaranteed present (Xvfb + Mesa), so a skip there means
// the very thing the test exists to prove (EGL config match, X11 setup) silently
// did not run. On a bare dev box the env var is unset and the test skips cleanly.
func requireOrSkip(t *testing.T, format string, args ...any) {
	t.Helper()
	if os.Getenv("POLYRED_REQUIRE_WINDOW") != "" {
		t.Fatalf("POLYRED_REQUIRE_WINDOW set but the windowed path is unavailable: "+format, args...)
	}
	t.Skipf(format, args...)
}

// TestX11WindowedPresent exercises the cgo-free X11 + EGL + GLES windowed-present
// path end to end: open an X display, create and map a window, create an EGL
// window surface bound to it, make a GLES context current, draw the same
// textured fullscreen quad that flush() presents every frame, and read the
// pixels back. It is the only runtime check of the purego X11 struct offsets,
// the EGL window-surface config match, and the gl.Functions marshaling
// (out-parameters, the **char of ShaderSource, the offset-as-uintptr of
// VertexAttribPointer, the pixel pointer of TexImage2D, floats), none of which a
// cgo-free *build* can prove.
//
// It runs under Xvfb + Mesa llvmpipe in CI (see the gl-probe workflow) with
// POLYRED_REQUIRE_WINDOW=1 so a skip is a failure there.
func TestX11WindowedPresent(t *testing.T) {
	if os.Getenv("DISPLAY") == "" {
		requireOrSkip(t, "no X display (set DISPLAY / run under Xvfb)")
	}
	// X11 + GL are thread-bound; pin this goroutine like run()/draw() do.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := loadX11(); err != nil {
		requireOrSkip(t, "libX11 unavailable: %v", err)
	}
	d, _, _ := purego.SyscallN(_XOpenDisplay, 0)
	if d == 0 {
		requireOrSkip(t, "XOpenDisplay returned NULL (no reachable X server)")
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
		requireOrSkip(t, "no EGL/GLES runtime (libEGL/libGLESv2/driver missing): %v", err)
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

	// Mirror draw()/flush(): upload a solid-red texture and draw the fullscreen
	// quad that samples it, then read it back. This drives the real present path
	// (BufferData, ShaderSource's **char, VertexAttribPointer's offset-as-uintptr,
	// TexImage2D's pixel pointer, DrawArrays) rather than only a clear. Red is
	// sRGB-invariant (0 and 1 map to themselves) and channel-specific, so a
	// channel-swapped readback is caught too.
	f.Viewport(0, 0, w, h)
	f.ClearColor(0, 0, 0, 1) // clear black so a passing red readback means the quad drew
	f.Clear(gl.COLOR_BUFFER_BIT)

	vertices := slice2bytes([]float32{
		-1, +1, 0, 0,
		+1, +1, 1, 0,
		-1, -1, 0, 1,
		+1, -1, 1, 1,
	})
	vbo := f.CreateBuffer()
	f.BindBuffer(gl.ARRAY_BUFFER, vbo)
	f.BufferData(gl.ARRAY_BUFFER, len(vertices), gl.STATIC_DRAW, vertices)
	defer f.DeleteBuffer(vbo)

	program, err := gl.CreateProgram(f, vert, frag, []string{"position", "uvcoord"})
	if err != nil {
		t.Fatalf("gl.CreateProgram (ShaderSource **char marshaling): %v", err)
	}
	f.UseProgram(program)
	defer f.DeleteProgram(program)

	position := f.GetAttribLocation(program, "position")
	uvcoord := f.GetAttribLocation(program, "uvcoord")
	f.EnableVertexAttribArray(position)
	f.EnableVertexAttribArray(uvcoord)
	f.VertexAttribPointer(position, 2, gl.FLOAT, false, 4*4, 0)
	f.VertexAttribPointer(uvcoord, 2, gl.FLOAT, false, 4*4, 2*4)

	// Solid opaque-red source image.
	src := make([]byte, w*h*4)
	for i := 0; i < len(src); i += 4 {
		src[i], src[i+1], src[i+2], src[i+3] = 255, 0, 0, 255
	}
	tex := f.CreateTexture()
	f.BindTexture(gl.TEXTURE_2D, tex)
	defer f.DeleteTexture(tex)
	f.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, w, h, gl.RGBA, gl.UNSIGNED_BYTE, src)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	f.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	f.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	f.Finish()

	pix := make([]byte, w*h*4)
	f.ReadPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, pix)
	// Sample the center pixel to avoid any edge interpolation.
	off := ((h/2)*w + w/2) * 4
	got := [4]byte{pix[off], pix[off+1], pix[off+2], pix[off+3]}
	want := [4]byte{255, 0, 0, 255}
	for i := range want {
		if diff := int(got[i]) - int(want[i]); diff < -2 || diff > 2 {
			t.Fatalf("textured-quad readback center = %v, want ~%v (gl present-path marshaling)", got, want)
		}
	}

	// Exercise eglSwapBuffers; it must not error on a mapped window surface.
	if err := ctx.Present(); err != nil {
		t.Fatalf("eglSwapBuffers failed: %v", err)
	}
}
