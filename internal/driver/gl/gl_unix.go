// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin || linux

package gl

/*
#cgo CFLAGS: -Werror

#cgo linux LDFLAGS: -lGL

#cgo darwin LDFLAGS: -framework OpenGL
#cgo darwin CFLAGS: -DGL_SILENCE_DEPRECATION

#include <stdlib.h>

#ifdef __APPLE__
	#include "TargetConditionals.h"
	#include <OpenGL/gl.h>
#else
#define __USE_GNU
#include <GL/gl.h>
#endif
*/
import "C"
import "unsafe"

func DrawBuffer(buf Enum) {
	C.glDrawBuffer(C.GLenum(buf))
}

func PixelZoom(xfactor, yfactor float32) {
	C.glPixelZoom(C.GLfloat(xfactor), C.GLfloat(yfactor))
}

func RasterPos2d(x, y float32) {
	C.glRasterPos2d(C.double(x), C.double(y))
}

func Viewport(x, y int, width, height int32) {
	C.glViewport(C.GLint(x), C.GLint(y), C.GLsizei(width), C.GLsizei(height))
}

func DrawPixels(width, height int32, format, xtype Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glDrawPixels(C.GLsizei(width), C.GLsizei(height), C.GLenum(format), C.GLenum(xtype), p)
}

func Clear(mask Enum) {
	C.glClear(C.GLbitfield(mask))
}

func ClearColor(red, green, blue, alpha float32) {
	C.glClearColor(C.GLfloat(red), C.GLfloat(green), C.GLfloat(blue), C.GLfloat(alpha))
}

func Finish() {
	C.glFinish()
}

func CreateProgram() Enum         { panic("unimplemented") }
func AttachShader() Enum          { panic("unimplemented") }
func LinkProgram() Enum           { panic("unimplemented") }
func CreateShader() Enum          { panic("unimplemented") }
func ShaderSource() Enum          { panic("unimplemented") }
func CompileShader() Enum         { panic("unimplemented") }
func GetShaderiv() Enum           { panic("unimplemented") }
func GetShaderInfoLog() Enum      { panic("unimplemented") }
func GetProgramiv() Enum          { panic("unimplemented") }
func GetProgramInfoLog() Enum     { panic("unimplemented") }
func DeleteShader() Enum          { panic("unimplemented") }
func CreateBuffer() Enum          { panic("unimplemented") }
func BindBuffer() Enum            { panic("unimplemented") }
func BufferData() Enum            { panic("unimplemented") }
func BindBufferBase() Enum        { panic("unimplemented") }
func DispatchCompute() Enum       { panic("unimplemented") }
func GetNamedBufferSubData() Enum { panic("unimplemented") }
func UseProgram() Enum            { panic("unimplemented") }
