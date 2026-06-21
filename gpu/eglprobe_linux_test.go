// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package gpu

import (
	"os"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
)

// EGL / GLES constants used by the probe.
const (
	eglDefaultDisplay = 0
	eglNoContext      = 0
	eglNoSurface      = 0
	eglOpenGLESAPI    = 0x30A0
	eglNone           = 0x3038
	eglContextMajor   = 0x3098 // EGL_CONTEXT_CLIENT_VERSION / MAJOR
	eglRenderableType = 0x3040
	eglOpenGLES3Bit   = 0x0040
	eglSurfaceType    = 0x3033
	eglPbufferBit     = 0x0001
	eglRedSize        = 0x3024
	eglGreenSize      = 0x3023
	eglBlueSize       = 0x3022

	glVersion                        = 0x1F02
	glRenderer                       = 0x1F01
	glMaxComputeWorkGroupInvocations = 0x90EB
)

// TestEGLSurfacelessGLESCompute proves, cgo-free (via purego), that a headless
// OpenGL ES 3.1 context with compute-shader support can be created on the CI
// runner using Mesa's surfaceless EGL platform (set EGL_PLATFORM=surfaceless).
// It is the foundation the cgo-free Linux GL backend (item #2) builds on; if it
// passes in CI, the GL compute backend is CI-verifiable, not hardware-gated.
//
// It skips when the EGL/GLES libraries are absent (e.g. a dev box without Mesa),
// so it is a no-op locally and only does real work on the Mesa-provisioned CI
// job.
func TestEGLSurfacelessGLESCompute(t *testing.T) {
	// Only run under the dedicated gl-probe workflow, which sets the surfaceless
	// platform. This keeps the test out of the main `go test ./...` Linux job
	// (which runs under Xvfb/X11 and may not have the GLES runtime installed).
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the headless GLES compute probe")
	}
	egl, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libEGL.so.1 not available: %v", err)
	}
	gles, err := purego.Dlopen("libGLESv2.so.2", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libGLESv2.so.2 not available: %v", err)
	}
	sym := func(h uintptr, name string) uintptr {
		p, err := purego.Dlsym(h, name)
		if err != nil {
			t.Fatalf("dlsym %s: %v", name, err)
		}
		return p
	}
	eglGetDisplay := sym(egl, "eglGetDisplay")
	eglInitialize := sym(egl, "eglInitialize")
	eglBindAPI := sym(egl, "eglBindAPI")
	eglChooseConfig := sym(egl, "eglChooseConfig")
	eglCreateContext := sym(egl, "eglCreateContext")
	eglMakeCurrent := sym(egl, "eglMakeCurrent")
	eglGetError := sym(egl, "eglGetError")
	glGetString := sym(gles, "glGetString")
	glGetIntegerv := sym(gles, "glGetIntegerv")

	eglErr := func() uintptr { e, _, _ := purego.SyscallN(eglGetError); return e }

	dpy, _, _ := purego.SyscallN(eglGetDisplay, uintptr(eglDefaultDisplay))
	if dpy == 0 {
		t.Fatalf("eglGetDisplay returned EGL_NO_DISPLAY (set EGL_PLATFORM=surfaceless?)")
	}

	var major, minor int32
	if r, _, _ := purego.SyscallN(eglInitialize, dpy, uintptr(unsafe.Pointer(&major)), uintptr(unsafe.Pointer(&minor))); r == 0 {
		t.Fatalf("eglInitialize failed: 0x%x", eglErr())
	}
	t.Logf("EGL %d.%d", major, minor)

	purego.SyscallN(eglBindAPI, uintptr(eglOpenGLESAPI))

	attribs := []int32{
		eglRenderableType, eglOpenGLES3Bit,
		eglSurfaceType, eglPbufferBit,
		eglRedSize, 8, eglGreenSize, 8, eglBlueSize, 8,
		eglNone,
	}
	var cfg uintptr
	var n int32
	if r, _, _ := purego.SyscallN(eglChooseConfig, dpy, uintptr(unsafe.Pointer(&attribs[0])), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&n))); r == 0 || n == 0 {
		t.Fatalf("eglChooseConfig failed: r=%d n=%d err=0x%x", 0, n, eglErr())
	}

	ctxAttribs := []int32{eglContextMajor, 3, eglNone}
	ctx, _, _ := purego.SyscallN(eglCreateContext, dpy, cfg, uintptr(eglNoContext), uintptr(unsafe.Pointer(&ctxAttribs[0])))
	if ctx == 0 {
		t.Fatalf("eglCreateContext failed: 0x%x", eglErr())
	}

	if r, _, _ := purego.SyscallN(eglMakeCurrent, dpy, uintptr(eglNoSurface), uintptr(eglNoSurface), ctx); r == 0 {
		t.Fatalf("eglMakeCurrent (surfaceless) failed: 0x%x", eglErr())
	}

	ver, _, _ := purego.SyscallN(glGetString, uintptr(glVersion))
	rend, _, _ := purego.SyscallN(glGetString, uintptr(glRenderer))
	t.Logf("GL_VERSION:  %s", cStr(ver))
	t.Logf("GL_RENDERER: %s", cStr(rend))

	var maxInv int32
	purego.SyscallN(glGetIntegerv, uintptr(glMaxComputeWorkGroupInvocations), uintptr(unsafe.Pointer(&maxInv)))
	t.Logf("GL_MAX_COMPUTE_WORK_GROUP_INVOCATIONS = %d", maxInv)
	if maxInv <= 0 {
		t.Fatalf("no compute-shader support: max work-group invocations = %d", maxInv)
	}
}

func cStr(p uintptr) string {
	if p == 0 {
		return "<nil>"
	}
	var b []byte
	for i := 0; ; i++ {
		c := *(*byte)(unsafe.Pointer(p + uintptr(i)))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	return string(b)
}
