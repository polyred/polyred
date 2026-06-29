// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.
// Modified from https://github.com/gioui/gio

package egl

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

// EGL handle types. EGL objects are opaque pointers; ints are 32-bit. Mirrors
// the windows variant so the cross-platform egl.go compiles against either.
type (
	_EGLint           int32
	_EGLDisplay       uintptr
	_EGLConfig        uintptr
	_EGLContext       uintptr
	_EGLSurface       uintptr
	NativeDisplayType uintptr
	NativeWindowType  uintptr
)

// Resolved libEGL entry points (purego function pointers).
var (
	_eglChooseConfig        uintptr
	_eglCreateContext       uintptr
	_eglCreateWindowSurface uintptr
	_eglDestroyContext      uintptr
	_eglDestroySurface      uintptr
	_eglGetConfigAttrib     uintptr
	_eglGetDisplay          uintptr
	_eglGetError            uintptr
	_eglInitialize          uintptr
	_eglMakeCurrent         uintptr
	_eglReleaseThread       uintptr
	_eglSwapInterval        uintptr
	_eglSwapBuffers         uintptr
	_eglTerminate           uintptr
	_eglQueryString         uintptr
	_eglWaitClient          uintptr
)

var loadOnce sync.Once

func loadEGL() error {
	var err error
	loadOnce.Do(func() {
		err = loadEGLSymbols()
	})
	return err
}

func loadEGLSymbols() error {
	h, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("egl: failed to load libEGL.so.1: %w", err)
	}
	var loadErr error
	sym := func(name string) uintptr {
		p, e := purego.Dlsym(h, name)
		if e != nil && loadErr == nil {
			loadErr = fmt.Errorf("egl: dlsym %s: %w", name, e)
		}
		return p
	}
	_eglChooseConfig = sym("eglChooseConfig")
	_eglCreateContext = sym("eglCreateContext")
	_eglCreateWindowSurface = sym("eglCreateWindowSurface")
	_eglDestroyContext = sym("eglDestroyContext")
	_eglDestroySurface = sym("eglDestroySurface")
	_eglGetConfigAttrib = sym("eglGetConfigAttrib")
	_eglGetDisplay = sym("eglGetDisplay")
	_eglGetError = sym("eglGetError")
	_eglInitialize = sym("eglInitialize")
	_eglMakeCurrent = sym("eglMakeCurrent")
	_eglReleaseThread = sym("eglReleaseThread")
	_eglSwapInterval = sym("eglSwapInterval")
	_eglSwapBuffers = sym("eglSwapBuffers")
	_eglTerminate = sym("eglTerminate")
	_eglQueryString = sym("eglQueryString")
	_eglWaitClient = sym("eglWaitClient")
	return loadErr
}

func eglChooseConfig(disp _EGLDisplay, attribs []_EGLint) (_EGLConfig, bool) {
	var cfg _EGLConfig
	var ncfg _EGLint
	r, _, _ := purego.SyscallN(_eglChooseConfig, uintptr(disp), uintptr(unsafe.Pointer(&attribs[0])), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&ncfg)))
	return cfg, r != 0
}

func eglCreateContext(disp _EGLDisplay, cfg _EGLConfig, shareCtx _EGLContext, attribs []_EGLint) _EGLContext {
	c, _, _ := purego.SyscallN(_eglCreateContext, uintptr(disp), uintptr(cfg), uintptr(shareCtx), uintptr(unsafe.Pointer(&attribs[0])))
	return _EGLContext(c)
}

func eglCreateWindowSurface(disp _EGLDisplay, conf _EGLConfig, win NativeWindowType, attribs []_EGLint) _EGLSurface {
	s, _, _ := purego.SyscallN(_eglCreateWindowSurface, uintptr(disp), uintptr(conf), uintptr(win), uintptr(unsafe.Pointer(&attribs[0])))
	return _EGLSurface(s)
}

func eglDestroySurface(disp _EGLDisplay, surf _EGLSurface) bool {
	r, _, _ := purego.SyscallN(_eglDestroySurface, uintptr(disp), uintptr(surf))
	return r != 0
}

func eglDestroyContext(disp _EGLDisplay, ctx _EGLContext) bool {
	r, _, _ := purego.SyscallN(_eglDestroyContext, uintptr(disp), uintptr(ctx))
	return r != 0
}

func eglGetConfigAttrib(disp _EGLDisplay, cfg _EGLConfig, attr _EGLint) (_EGLint, bool) {
	var val _EGLint
	r, _, _ := purego.SyscallN(_eglGetConfigAttrib, uintptr(disp), uintptr(cfg), uintptr(attr), uintptr(unsafe.Pointer(&val)))
	return val, r != 0
}

func eglGetError() _EGLint {
	e, _, _ := purego.SyscallN(_eglGetError)
	return _EGLint(e)
}

func eglInitialize(disp _EGLDisplay) (_EGLint, _EGLint, bool) {
	var maj, min _EGLint
	r, _, _ := purego.SyscallN(_eglInitialize, uintptr(disp), uintptr(unsafe.Pointer(&maj)), uintptr(unsafe.Pointer(&min)))
	return maj, min, r != 0
}

func eglMakeCurrent(disp _EGLDisplay, draw, read _EGLSurface, ctx _EGLContext) bool {
	r, _, _ := purego.SyscallN(_eglMakeCurrent, uintptr(disp), uintptr(draw), uintptr(read), uintptr(ctx))
	return r != 0
}

func eglReleaseThread() bool {
	r, _, _ := purego.SyscallN(_eglReleaseThread)
	return r != 0
}

func eglSwapBuffers(disp _EGLDisplay, surf _EGLSurface) bool {
	r, _, _ := purego.SyscallN(_eglSwapBuffers, uintptr(disp), uintptr(surf))
	return r != 0
}

func eglSwapInterval(disp _EGLDisplay, interval _EGLint) bool {
	r, _, _ := purego.SyscallN(_eglSwapInterval, uintptr(disp), uintptr(interval))
	return r != 0
}

func eglTerminate(disp _EGLDisplay) bool {
	r, _, _ := purego.SyscallN(_eglTerminate, uintptr(disp))
	return r != 0
}

func eglQueryString(disp _EGLDisplay, name _EGLint) string {
	r, _, _ := purego.SyscallN(_eglQueryString, uintptr(disp), uintptr(name))
	return goString(r)
}

func eglGetDisplay(disp NativeDisplayType) _EGLDisplay {
	d, _, _ := purego.SyscallN(_eglGetDisplay, uintptr(disp))
	return _EGLDisplay(d)
}

func eglWaitClient() bool {
	r, _, _ := purego.SyscallN(_eglWaitClient)
	return r != 0
}

// goString converts a NUL-terminated C string pointer to a Go string.
func goString(p uintptr) string {
	if p == 0 {
		return ""
	}
	a := (*[1 << 20]byte)(unsafe.Pointer(p))
	i := 0
	for a[i] != 0 {
		i++
	}
	return string(a[:i])
}
