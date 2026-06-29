// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package gpu

import "github.com/ebitengine/purego"

// Shared-library names for the EGL/GLES entry points on Linux (Mesa).
const (
	eglLibName  = "libEGL.so.1"
	glesLibName = "libGLESv2.so.2"
)

// glDlopen loads a shared library by name. On Linux it goes through purego's
// dlopen with RTLD_NOW|RTLD_GLOBAL (unchanged from the original inline calls).
func glDlopen(name string) (uintptr, error) {
	return purego.Dlopen(name, purego.RTLD_NOW|purego.RTLD_GLOBAL)
}

// glDlsym resolves a symbol in a library handle returned by glDlopen.
func glDlsym(handle uintptr, name string) (uintptr, error) {
	return purego.Dlsym(handle, name)
}
