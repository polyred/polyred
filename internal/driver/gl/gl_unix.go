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

func CreateProgram() Program {
	return Program(C.glCreateProgram())
}

func CreateShader(typ Enum) Shader {
	return Shader(C.glCreateShader(C.GLenum(typ)))
}

func AttachShader(p Program, s Shader) {
	C.glAttachShader(C.GLuint(p), C.GLuint(s))
}

func LinkProgram(p Program) {
	C.glLinkProgram(C.GLuint(p))
}

func ShaderSource(s Shader, src string) {
	csrc := C.CString(src)
	defer C.free(unsafe.Pointer(csrc))
	strlen := C.GLint(len(src))
	C.glShaderSource(C.GLuint(s), 1, &csrc, &strlen)
}

func CompileShader(s Shader) {
	C.glCompileShader(C.GLuint(s))
}

func GetShaderiv(s Shader, pname Enum) int {
	var x C.GLint
	C.glGetShaderiv(C.GLuint(s), C.GLenum(pname), &x)
	return int(x)
}

func GetShaderInfoLog(s Shader) string {
	n := GetShaderiv(s, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetShaderInfoLog(C.GLuint(s), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func GetProgramiv(p Program, pname Enum) int {
	var x C.GLint
	C.glGetProgramiv(C.GLuint(p), C.GLenum(pname), &x)
	return int(x)
}

func GetProgramInfoLog(p Program) string {
	n := GetProgramiv(p, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	C.glGetProgramInfoLog(C.GLuint(p), C.GLsizei(len(buf)), nil, (*C.GLchar)(unsafe.Pointer(&buf[0])))
	return string(buf)
}

func DeleteShader(s Shader) {
	C.glDeleteShader(C.GLuint(s))
}

func CreateBuffer() Buffer {
	var x C.GLuint
	C.glGenBuffers(1, &x)
	return Buffer(x)
}

func BindBuffer(target Enum, b Buffer) {
	C.glBindBuffer(C.GLenum(target), C.GLuint(b))
}

func BufferData(target Enum, size int, usage Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	C.glBufferData(C.GLenum(target), C.GLsizeiptr(size), p, C.GLenum(usage))
}

func BindBufferBase(target Enum, index int, b Buffer) {
	C.glBindBufferBase(C.GLenum(target), C.GLuint(index), C.GLuint(b))
}

func DispatchCompute(x, y, z int) {
	C.glDispatchCompute(C.GLuint(x), C.GLuint(y), C.GLuint(z))
}

// func GetNamedBufferSubData() Enum { panic("unimplemented") }

func UseProgram(p Program) {
	C.glUseProgram(C.GLuint(p))
}
