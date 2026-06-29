// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build windows

package gpu

import "syscall"

// Shared-library names for the EGL/GLES entry points on Windows (ANGLE).
const (
	eglLibName  = "libEGL.dll"
	glesLibName = "libGLESv2.dll"
)

// glDlopen loads a DLL by name. purego's dlopen is Unix-only, so on Windows we
// use the loader directly; the PE loader resolves each DLL's static imports
// (libEGL.dll pulls in libGLESv2.dll) and ANGLE loads d3dcompiler_47.dll itself
// from the standard search path. The returned handle feeds glDlsym, and the GL
// call sites stay on purego.SyscallN (available on windows/amd64).
func glDlopen(name string) (uintptr, error) {
	h, err := syscall.LoadLibrary(name)
	return uintptr(h), err
}

// glDlsym resolves a symbol in a DLL handle returned by glDlopen.
func glDlsym(handle uintptr, name string) (uintptr, error) {
	return syscall.GetProcAddress(syscall.Handle(handle), name)
}
