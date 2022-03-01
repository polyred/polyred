// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build windows

package gl

import (
	"math"
	"runtime"
	"syscall"
	"unsafe"
)

var (
	opengl32       = syscall.NewLazyDLL("opengl32.dll")
	_glMakeCurrent = opengl32.NewProc("wglMakeCurrent")
	_glDrawBuffer  = opengl32.NewProc("glDrawBuffer")
	_glPixelZoom   = opengl32.NewProc("glPixelZoom")
	_glRasterPos2d = opengl32.NewProc("glRasterPos2d")
	_glViewport    = opengl32.NewProc("glViewport")
	_glDrawPixels  = opengl32.NewProc("glDrawPixels")
	_glFinish      = opengl32.NewProc("glFinish")
)

func MakeCurrent(hdc syscall.Handle) {
	syscall.Syscall(_glMakeCurrent.Addr(), 2, uintptr(hdc), 0, 0)
}

func DrawBuffer(buf Enum) {
	syscall.Syscall(_glDrawBuffer.Addr(), 1, uintptr(buf), 0, 0)
}

func PixelZoom(xfactor, yfactor float32) {
	syscall.Syscall(_glPixelZoom.Addr(), 2, uintptr(math.Float32bits(xfactor)), uintptr(math.Float32bits(yfactor)), 0)
}

func RasterPos2d(x, y float64) {
	syscall.Syscall(_glRasterPos2d.Addr(), 2, uintptr(math.Float64bits(x)), uintptr(math.Float64bits(y)), 0)
}

func Viewport(x, y int32, width, height int32) {
	syscall.Syscall6(_glViewport.Addr(), 4, uintptr(x), uintptr(y), uintptr(width), uintptr(height), 0, 0)
}

func DrawPixels(width, height int32, format, xtype Enum, data []byte) {

	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}

	syscall.Syscall6(_glDrawPixels.Addr(), 5, uintptr(width), uintptr(height), uintptr(format), uintptr(xtype), uintptr(p), 0)
	runtime.KeepAlive(p) // See https://golang.org/issue/34474
}

func Finish() {
	syscall.Syscall(_glFinish.Addr(), 0, 0, 0, 0)
}
