// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build linux

package egl

/*
#cgo LDFLAGS: -ldl

#include <dlfcn.h>
#include <X11/Xlib.h>

void*    libX11;
Display* (*X11OpenDisplay)(int);
void     (*X11CloseDisplay)(Display*);

int initX11() {
	libX11 = dlopen("libX11.so", RTLD_LAZY);
	if (!libX11) {
		return -1;
	}
	X11OpenDisplay = (Display* (*)(int)) dlsym(libX11, "XOpenDisplay");
	X11CloseDisplay = (void (*)(Display*)) dlsym(libX11, "XCloseDisplay");

	return 0;
}

Display* OpenDisplay(int n) {
	return (*X11OpenDisplay)(n);
}

void CloseDisplay(Display* display) {
	(*X11CloseDisplay)(display);
}
*/
import "C"
import "runtime"

func init() {
	if int(C.initX11()) != 0 {
		panic("egl: cannot initialize X11 from libX11.so")
	}
}

func NewDisplay() NativeDisplayType {
	display := C.OpenDisplay(0)
	if display == nil {
		panic("x11: cannot connect to the X server")
	}
	runtime.SetFinalizer(display, func(display *C.Display) {

	})

	return NativeDisplayType(display)
}
