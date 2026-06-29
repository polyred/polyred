// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file is inspired by https://gioui.org

//go:build darwin || linux

package gl

import (
	"fmt"
	"runtime"
	"strings"
	"unsafe"

	"github.com/ebitengine/purego"
)

type Context any

type Functions struct {
	// Query caches.
	uints  [100]uint32
	ints   [100]int32
	floats [100]float32

	// GL ES 2.0 functions.
	glActiveTexture                       func(texture uint32)
	glAttachShader                        func(program, shader uint32)
	glBindAttribLocation                  func(program, index uint32, name unsafe.Pointer)
	glBindBuffer                          func(target, buffer uint32)
	glBindFramebuffer                     func(target, framebuffer uint32)
	glBindRenderbuffer                    func(target, renderbuffer uint32)
	glBindTexture                         func(target, texture uint32)
	glBlendEquation                       func(mode uint32)
	glBlendFuncSeparate                   func(srcRGB, dstRGB, srcA, dstA uint32)
	glBufferData                          func(target uint32, size int, data unsafe.Pointer, usage uint32)
	glBufferSubData                       func(target uint32, offset int, size int, data unsafe.Pointer)
	glCheckFramebufferStatus              func(target uint32) uint32
	glClear                               func(mask uint32)
	glClearColor                          func(red, green, blue, alpha float32)
	glClearDepthf                         func(d float32)
	glCompileShader                       func(shader uint32)
	glCopyTexSubImage2D                   func(target uint32, level, xoffset, yoffset, x, y, width, height int32)
	glCreateProgram                       func() uint32
	glCreateShader                        func(ty uint32) uint32
	glDeleteBuffers                       func(n int32, buffers unsafe.Pointer)
	glDeleteFramebuffers                  func(n int32, framebuffers unsafe.Pointer)
	glDeleteProgram                       func(program uint32)
	glDeleteRenderbuffers                 func(n int32, renderbuffers unsafe.Pointer)
	glDeleteShader                        func(shader uint32)
	glDeleteTextures                      func(n int32, textures unsafe.Pointer)
	glDepthFunc                           func(fn uint32)
	glDepthMask                           func(flag uint8)
	glDisable                             func(cap uint32)
	glDisableVertexAttribArray            func(index uint32)
	glDrawArrays                          func(mode uint32, first, count int32)
	glDrawElements                        func(mode uint32, count int32, ty uint32, indices uintptr)
	glEnable                              func(cap uint32)
	glEnableVertexAttribArray             func(index uint32)
	glFinish                              func()
	glFlush                               func()
	glFramebufferRenderbuffer             func(target, attachment, renderbuffertarget, renderbuffer uint32)
	glFramebufferTexture2D                func(target, attachment, textarget, texture uint32, level int32)
	glGenBuffers                          func(n int32, buffers unsafe.Pointer)
	glGenFramebuffers                     func(n int32, framebuffers unsafe.Pointer)
	glGenRenderbuffers                    func(n int32, renderbuffers unsafe.Pointer)
	glGenTextures                         func(n int32, textures unsafe.Pointer)
	glGetError                            func() uint32
	glGetFramebufferAttachmentParameteriv func(target, attachment, pname uint32, params unsafe.Pointer)
	glGetFloatv                           func(pname uint32, data unsafe.Pointer)
	glGetIntegerv                         func(pname uint32, data unsafe.Pointer)
	glGetProgramiv                        func(program, pname uint32, params unsafe.Pointer)
	glGetProgramInfoLog                   func(program uint32, bufSize int32, length, infoLog unsafe.Pointer)
	glGetRenderbufferParameteriv          func(target, pname uint32, params unsafe.Pointer)
	glGetShaderiv                         func(shader, pname uint32, params unsafe.Pointer)
	glGetShaderInfoLog                    func(shader uint32, bufSize int32, length, infoLog unsafe.Pointer)
	glGetString                           func(name uint32) *byte
	glGetUniformLocation                  func(program uint32, name unsafe.Pointer) int32
	glGetAttribLocation                   func(program uint32, name unsafe.Pointer) int32
	glGetVertexAttribiv                   func(index, pname uint32, params unsafe.Pointer)
	glGetVertexAttribPointerv             func(index, pname uint32, params unsafe.Pointer)
	glIsEnabled                           func(cap uint32) uint8
	glLinkProgram                         func(program uint32)
	glPixelStorei                         func(pname uint32, param int32)
	glReadPixels                          func(x, y, width, height int32, format, ty uint32, pixels unsafe.Pointer)
	glRenderbufferStorage                 func(target, internalformat uint32, width, height int32)
	glScissor                             func(x, y, width, height int32)
	glShaderSource                        func(shader uint32, count int32, str, length unsafe.Pointer)
	glTexImage2D                          func(target uint32, level, internalformat, width, height, border int32, format, ty uint32, pixels unsafe.Pointer)
	glTexParameteri                       func(target, pname uint32, param int32)
	glTexSubImage2D                       func(target uint32, level, xoffset, yoffset, width, height int32, format, ty uint32, pixels unsafe.Pointer)
	glUniform1f                           func(location int32, v0 float32)
	glUniform1i                           func(location, v0 int32)
	glUniform2f                           func(location int32, v0, v1 float32)
	glUniform3f                           func(location int32, v0, v1, v2 float32)
	glUniform4f                           func(location int32, v0, v1, v2, v3 float32)
	glUseProgram                          func(program uint32)
	glVertexAttribPointer                 func(index uint32, size int32, ty uint32, normalized uint8, stride int32, pointer uintptr)
	glViewport                            func(x, y, width, height int32)

	// Extensions and GL ES 3 functions.
	glBindBufferBase        func(target, index, buffer uint32)
	glBindVertexArray       func(array uint32)
	glGetIntegeri_v         func(pname, idx uint32, data unsafe.Pointer)
	glGetUniformBlockIndex  func(program uint32, uniformBlockName unsafe.Pointer) uint32
	glUniformBlockBinding   func(program, uniformBlockIndex, uniformBlockBinding uint32)
	glInvalidateFramebuffer func(target uint32, numAttachments int32, attachments unsafe.Pointer)
	glGetStringi            func(name, index uint32) *byte
	glBeginQuery            func(target, id uint32)
	glDeleteQueries         func(n int32, ids unsafe.Pointer)
	glEndQuery              func(target uint32)
	glGenQueries            func(n int32, ids unsafe.Pointer)
	glGetQueryObjectuiv     func(id, pname uint32, params unsafe.Pointer)
	glDeleteVertexArrays    func(n int32, ids unsafe.Pointer)
	glGenVertexArrays       func(n int32, ids unsafe.Pointer)
	glMemoryBarrier         func(barriers uint32)
	glDispatchCompute       func(x, y, z uint32)
	glMapBufferRange        func(target uint32, offset, length int, access uint32) unsafe.Pointer
	glUnmapBuffer           func(target uint32) uint8
	glBindImageTexture      func(unit, texture uint32, level int32, layered uint8, layer int32, access, format uint32)
	glTexStorage2D          func(target uint32, levels int32, internalformat uint32, width, height int32)
	glBlitFramebuffer       func(srcX0, srcY0, srcX1, srcY1, dstX0, dstY0, dstX1, dstY1 int32, mask, filter uint32)
	glGetProgramBinary      func(program uint32, bufsize int32, length, binaryFormat, binary unsafe.Pointer)
}

func NewFunctions() (*Functions, error) {
	f := new(Functions)
	err := f.load()
	if err != nil {
		return nil, err
	}
	return f, nil
}

func (f *Functions) Available() bool { return f != nil }

// goString converts a NUL-terminated C string to a Go string.
func goString(p *byte) string {
	if p == nil {
		return ""
	}
	var b strings.Builder
	for ptr := unsafe.Pointer(p); ; ptr = unsafe.Add(ptr, 1) {
		c := *(*byte)(ptr)
		if c == 0 {
			break
		}
		b.WriteByte(c)
	}
	return b.String()
}

func (f *Functions) load() error {
	var (
		loadErr  error
		libNames []string
		handles  []uintptr
	)
	switch runtime.GOOS {
	case "darwin":
		libNames = []string{"libGLESv2.dylib"}
	case "ios":
		libNames = []string{"/System/Library/Frameworks/OpenGLES.framework/OpenGLES"}
	case "android":
		libNames = []string{"libGLESv2.so", "libGLESv3.so"}
	default:
		libNames = []string{"libGLESv2.so.2"}
	}
	for _, lib := range libNames {
		if h, err := purego.Dlopen(lib, purego.RTLD_NOW|purego.RTLD_LOCAL); err == nil && h != 0 {
			handles = append(handles, h)
		}
	}
	if len(handles) == 0 {
		return fmt.Errorf("gl: no OpenGL implementation could be loaded (tried %q)", libNames)
	}
	// load returns the first matching symbol across the loaded handles, or 0.
	load := func(s string) uintptr {
		for _, h := range handles {
			if p, err := purego.Dlsym(h, s); err == nil && p != 0 {
				return p
			}
		}
		return 0
	}
	// must is like load but records an error if the symbol is missing.
	must := func(s string) uintptr {
		p := load(s)
		if p == 0 && loadErr == nil {
			loadErr = fmt.Errorf("gl: failed to load symbol %q", s)
		}
		return p
	}
	// reg registers the symbol as a typed Go func only if present. This avoids
	// purego.RegisterFunc panicking on a zero function pointer; missing required
	// symbols are reported through loadErr instead.
	reg := func(fptr any, p uintptr) {
		if p != 0 {
			purego.RegisterFunc(fptr, p)
		}
	}

	// GL ES 2.0 functions.
	reg(&f.glActiveTexture, must("glActiveTexture"))
	reg(&f.glAttachShader, must("glAttachShader"))
	reg(&f.glBindAttribLocation, must("glBindAttribLocation"))
	reg(&f.glBindBuffer, must("glBindBuffer"))
	reg(&f.glBindFramebuffer, must("glBindFramebuffer"))
	reg(&f.glBindRenderbuffer, must("glBindRenderbuffer"))
	reg(&f.glBindTexture, must("glBindTexture"))
	reg(&f.glBlendEquation, must("glBlendEquation"))
	reg(&f.glBlendFuncSeparate, must("glBlendFuncSeparate"))
	reg(&f.glBufferData, must("glBufferData"))
	reg(&f.glBufferSubData, must("glBufferSubData"))
	reg(&f.glCheckFramebufferStatus, must("glCheckFramebufferStatus"))
	reg(&f.glClear, must("glClear"))
	reg(&f.glClearColor, must("glClearColor"))
	reg(&f.glClearDepthf, must("glClearDepthf"))
	reg(&f.glCompileShader, must("glCompileShader"))
	reg(&f.glCopyTexSubImage2D, must("glCopyTexSubImage2D"))
	reg(&f.glCreateProgram, must("glCreateProgram"))
	reg(&f.glCreateShader, must("glCreateShader"))
	reg(&f.glDeleteBuffers, must("glDeleteBuffers"))
	reg(&f.glDeleteFramebuffers, must("glDeleteFramebuffers"))
	reg(&f.glDeleteProgram, must("glDeleteProgram"))
	reg(&f.glDeleteRenderbuffers, must("glDeleteRenderbuffers"))
	reg(&f.glDeleteShader, must("glDeleteShader"))
	reg(&f.glDeleteTextures, must("glDeleteTextures"))
	reg(&f.glDepthFunc, must("glDepthFunc"))
	reg(&f.glDepthMask, must("glDepthMask"))
	reg(&f.glDisable, must("glDisable"))
	reg(&f.glDisableVertexAttribArray, must("glDisableVertexAttribArray"))
	reg(&f.glDrawArrays, must("glDrawArrays"))
	reg(&f.glDrawElements, must("glDrawElements"))
	reg(&f.glEnable, must("glEnable"))
	reg(&f.glEnableVertexAttribArray, must("glEnableVertexAttribArray"))
	reg(&f.glFinish, must("glFinish"))
	reg(&f.glFlush, must("glFlush"))
	reg(&f.glFramebufferRenderbuffer, must("glFramebufferRenderbuffer"))
	reg(&f.glFramebufferTexture2D, must("glFramebufferTexture2D"))
	reg(&f.glGenBuffers, must("glGenBuffers"))
	reg(&f.glGenFramebuffers, must("glGenFramebuffers"))
	reg(&f.glGenRenderbuffers, must("glGenRenderbuffers"))
	reg(&f.glGenTextures, must("glGenTextures"))
	reg(&f.glGetError, must("glGetError"))
	reg(&f.glGetFramebufferAttachmentParameteriv, must("glGetFramebufferAttachmentParameteriv"))
	reg(&f.glGetIntegerv, must("glGetIntegerv"))
	reg(&f.glGetFloatv, must("glGetFloatv"))
	reg(&f.glGetProgramiv, must("glGetProgramiv"))
	reg(&f.glGetProgramInfoLog, must("glGetProgramInfoLog"))
	reg(&f.glGetRenderbufferParameteriv, must("glGetRenderbufferParameteriv"))
	reg(&f.glGetShaderiv, must("glGetShaderiv"))
	reg(&f.glGetShaderInfoLog, must("glGetShaderInfoLog"))
	reg(&f.glGetString, must("glGetString"))
	reg(&f.glGetUniformLocation, must("glGetUniformLocation"))
	reg(&f.glGetAttribLocation, must("glGetAttribLocation"))
	reg(&f.glGetVertexAttribiv, must("glGetVertexAttribiv"))
	reg(&f.glGetVertexAttribPointerv, must("glGetVertexAttribPointerv"))
	reg(&f.glIsEnabled, must("glIsEnabled"))
	reg(&f.glLinkProgram, must("glLinkProgram"))
	reg(&f.glPixelStorei, must("glPixelStorei"))
	reg(&f.glReadPixels, must("glReadPixels"))
	reg(&f.glRenderbufferStorage, must("glRenderbufferStorage"))
	reg(&f.glScissor, must("glScissor"))
	reg(&f.glShaderSource, must("glShaderSource"))
	reg(&f.glTexImage2D, must("glTexImage2D"))
	reg(&f.glTexParameteri, must("glTexParameteri"))
	reg(&f.glTexSubImage2D, must("glTexSubImage2D"))
	reg(&f.glUniform1f, must("glUniform1f"))
	reg(&f.glUniform1i, must("glUniform1i"))
	reg(&f.glUniform2f, must("glUniform2f"))
	reg(&f.glUniform3f, must("glUniform3f"))
	reg(&f.glUniform4f, must("glUniform4f"))
	reg(&f.glUseProgram, must("glUseProgram"))
	reg(&f.glVertexAttribPointer, must("glVertexAttribPointer"))
	reg(&f.glViewport, must("glViewport"))

	// Extensions and GL ES 3 functions.
	reg(&f.glBindBufferBase, load("glBindBufferBase"))
	reg(&f.glBindVertexArray, load("glBindVertexArray"))
	reg(&f.glGetIntegeri_v, load("glGetIntegeri_v"))
	reg(&f.glGetUniformBlockIndex, load("glGetUniformBlockIndex"))
	reg(&f.glUniformBlockBinding, load("glUniformBlockBinding"))
	reg(&f.glGetStringi, load("glGetStringi"))

	// Framebuffer invalidation falls back to EXT_discard_framebuffer.
	if p := load("glInvalidateFramebuffer"); p != 0 {
		reg(&f.glInvalidateFramebuffer, p)
	} else {
		reg(&f.glInvalidateFramebuffer, load("glDiscardFramebufferEXT"))
	}

	if p := load("glBeginQuery"); p != 0 {
		reg(&f.glBeginQuery, p)
	} else {
		reg(&f.glBeginQuery, load("glBeginQueryEXT"))
	}
	if p := load("glDeleteQueries"); p != 0 {
		reg(&f.glDeleteQueries, p)
	} else {
		reg(&f.glDeleteQueries, load("glDeleteQueriesEXT"))
	}
	if p := load("glEndQuery"); p != 0 {
		reg(&f.glEndQuery, p)
	} else {
		reg(&f.glEndQuery, load("glEndQueryEXT"))
	}
	if p := load("glGenQueries"); p != 0 {
		reg(&f.glGenQueries, p)
	} else {
		reg(&f.glGenQueries, load("glGenQueriesEXT"))
	}
	if p := load("glGetQueryObjectuiv"); p != 0 {
		reg(&f.glGetQueryObjectuiv, p)
	} else {
		reg(&f.glGetQueryObjectuiv, load("glGetQueryObjectuivEXT"))
	}

	reg(&f.glDeleteVertexArrays, load("glDeleteVertexArrays"))
	reg(&f.glGenVertexArrays, load("glGenVertexArrays"))
	reg(&f.glMemoryBarrier, load("glMemoryBarrier"))
	reg(&f.glDispatchCompute, load("glDispatchCompute"))
	reg(&f.glMapBufferRange, load("glMapBufferRange"))
	reg(&f.glUnmapBuffer, load("glUnmapBuffer"))
	reg(&f.glBindImageTexture, load("glBindImageTexture"))
	reg(&f.glTexStorage2D, load("glTexStorage2D"))
	reg(&f.glBlitFramebuffer, load("glBlitFramebuffer"))
	reg(&f.glGetProgramBinary, load("glGetProgramBinary"))

	return loadErr
}

func (f *Functions) ActiveTexture(texture Enum) {
	f.glActiveTexture(uint32(texture))
}

func (f *Functions) AttachShader(p Program, s Shader) {
	f.glAttachShader(uint32(p.V), uint32(s.V))
}

func (f *Functions) BeginQuery(target Enum, query Query) {
	f.glBeginQuery(uint32(target), uint32(query.V))
}

func (f *Functions) BindAttribLocation(p Program, a Attrib, name string) {
	b := cstr(name)
	f.glBindAttribLocation(uint32(p.V), uint32(a), unsafe.Pointer(&b[0]))
	runtime.KeepAlive(b)
}

func (f *Functions) BindBufferBase(target Enum, index int, b Buffer) {
	f.glBindBufferBase(uint32(target), uint32(index), uint32(b.V))
}

func (f *Functions) BindBuffer(target Enum, b Buffer) {
	f.glBindBuffer(uint32(target), uint32(b.V))
}

func (f *Functions) BindFramebuffer(target Enum, fb Framebuffer) {
	f.glBindFramebuffer(uint32(target), uint32(fb.V))
}

func (f *Functions) BindRenderbuffer(target Enum, fb Renderbuffer) {
	f.glBindRenderbuffer(uint32(target), uint32(fb.V))
}

func (f *Functions) BindImageTexture(unit int, t Texture, level int, layered bool, layer int, access, format Enum) {
	l := uint8(FALSE)
	if layered {
		l = uint8(TRUE)
	}
	f.glBindImageTexture(uint32(unit), uint32(t.V), int32(level), l, int32(layer), uint32(access), uint32(format))
}

func (f *Functions) BindTexture(target Enum, t Texture) {
	f.glBindTexture(uint32(target), uint32(t.V))
}

func (f *Functions) BindVertexArray(a VertexArray) {
	f.glBindVertexArray(uint32(a.V))
}

func (f *Functions) BlendEquation(mode Enum) {
	f.glBlendEquation(uint32(mode))
}

func (f *Functions) BlendFuncSeparate(srcRGB, dstRGB, srcA, dstA Enum) {
	f.glBlendFuncSeparate(uint32(srcRGB), uint32(dstRGB), uint32(srcA), uint32(dstA))
}

func (f *Functions) BlitFramebuffer(sx0, sy0, sx1, sy1, dx0, dy0, dx1, dy1 int, mask Enum, filter Enum) {
	f.glBlitFramebuffer(
		int32(sx0), int32(sy0), int32(sx1), int32(sy1),
		int32(dx0), int32(dy0), int32(dx1), int32(dy1),
		uint32(mask), uint32(filter),
	)
}

func (f *Functions) BufferData(target Enum, size int, usage Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	f.glBufferData(uint32(target), size, p, uint32(usage))
}

func (f *Functions) BufferSubData(target Enum, offset int, src []byte) {
	var p unsafe.Pointer
	if len(src) > 0 {
		p = unsafe.Pointer(&src[0])
	}
	f.glBufferSubData(uint32(target), offset, len(src), p)
}

func (f *Functions) CheckFramebufferStatus(target Enum) Enum {
	return Enum(f.glCheckFramebufferStatus(uint32(target)))
}

func (f *Functions) Clear(mask Enum) {
	f.glClear(uint32(mask))
}

func (f *Functions) ClearColor(red float32, green float32, blue float32, alpha float32) {
	f.glClearColor(red, green, blue, alpha)
}

func (f *Functions) ClearDepthf(d float32) {
	f.glClearDepthf(d)
}

func (f *Functions) CompileShader(s Shader) {
	f.glCompileShader(uint32(s.V))
}

func (f *Functions) CopyTexSubImage2D(target Enum, level, xoffset, yoffset, x, y, width, height int) {
	f.glCopyTexSubImage2D(uint32(target), int32(level), int32(xoffset), int32(yoffset), int32(x), int32(y), int32(width), int32(height))
}

func (f *Functions) CreateBuffer() Buffer {
	f.glGenBuffers(1, unsafe.Pointer(&f.uints[0]))
	return Buffer{uint(f.uints[0])}
}

func (f *Functions) CreateFramebuffer() Framebuffer {
	f.glGenFramebuffers(1, unsafe.Pointer(&f.uints[0]))
	return Framebuffer{uint(f.uints[0])}
}

func (f *Functions) CreateProgram() Program {
	return Program{uint(f.glCreateProgram())}
}

func (f *Functions) CreateQuery() Query {
	f.glGenQueries(1, unsafe.Pointer(&f.uints[0]))
	return Query{uint(f.uints[0])}
}

func (f *Functions) CreateRenderbuffer() Renderbuffer {
	f.glGenRenderbuffers(1, unsafe.Pointer(&f.uints[0]))
	return Renderbuffer{uint(f.uints[0])}
}

func (f *Functions) CreateShader(ty Enum) Shader {
	return Shader{uint(f.glCreateShader(uint32(ty)))}
}

func (f *Functions) CreateTexture() Texture {
	f.glGenTextures(1, unsafe.Pointer(&f.uints[0]))
	return Texture{uint(f.uints[0])}
}

func (f *Functions) CreateVertexArray() VertexArray {
	f.glGenVertexArrays(1, unsafe.Pointer(&f.uints[0]))
	return VertexArray{uint(f.uints[0])}
}

func (f *Functions) DeleteBuffer(v Buffer) {
	f.uints[0] = uint32(v.V)
	f.glDeleteBuffers(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DeleteFramebuffer(v Framebuffer) {
	f.uints[0] = uint32(v.V)
	f.glDeleteFramebuffers(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DeleteProgram(p Program) {
	f.glDeleteProgram(uint32(p.V))
}

func (f *Functions) DeleteQuery(query Query) {
	f.uints[0] = uint32(query.V)
	f.glDeleteQueries(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DeleteVertexArray(array VertexArray) {
	f.uints[0] = uint32(array.V)
	f.glDeleteVertexArrays(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DeleteRenderbuffer(v Renderbuffer) {
	f.uints[0] = uint32(v.V)
	f.glDeleteRenderbuffers(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DeleteShader(s Shader) {
	f.glDeleteShader(uint32(s.V))
}

func (f *Functions) DeleteTexture(v Texture) {
	f.uints[0] = uint32(v.V)
	f.glDeleteTextures(1, unsafe.Pointer(&f.uints[0]))
}

func (f *Functions) DepthFunc(v Enum) {
	f.glDepthFunc(uint32(v))
}

func (f *Functions) DepthMask(mask bool) {
	m := uint8(FALSE)
	if mask {
		m = uint8(TRUE)
	}
	f.glDepthMask(m)
}

func (f *Functions) DisableVertexAttribArray(a Attrib) {
	f.glDisableVertexAttribArray(uint32(a))
}

func (f *Functions) Disable(cap Enum) {
	f.glDisable(uint32(cap))
}

func (f *Functions) DrawArrays(mode Enum, first int, count int) {
	f.glDrawArrays(uint32(mode), int32(first), int32(count))
}

func (f *Functions) DrawElements(mode Enum, count int, ty Enum, offset int) {
	f.glDrawElements(uint32(mode), int32(count), uint32(ty), uintptr(offset))
}

func (f *Functions) DispatchCompute(x, y, z int) {
	f.glDispatchCompute(uint32(x), uint32(y), uint32(z))
}

func (f *Functions) Enable(cap Enum) {
	f.glEnable(uint32(cap))
}

func (f *Functions) EndQuery(target Enum) {
	f.glEndQuery(uint32(target))
}

func (f *Functions) EnableVertexAttribArray(a Attrib) {
	f.glEnableVertexAttribArray(uint32(a))
}

func (f *Functions) Finish() {
	f.glFinish()
}

func (f *Functions) Flush() {
	f.glFlush()
}

func (f *Functions) FramebufferRenderbuffer(target, attachment, renderbuffertarget Enum, renderbuffer Renderbuffer) {
	f.glFramebufferRenderbuffer(uint32(target), uint32(attachment), uint32(renderbuffertarget), uint32(renderbuffer.V))
}

func (f *Functions) FramebufferTexture2D(target, attachment, texTarget Enum, t Texture, level int) {
	f.glFramebufferTexture2D(uint32(target), uint32(attachment), uint32(texTarget), uint32(t.V), int32(level))
}

func (c *Functions) GetBinding(pname Enum) Object {
	return Object{uint(c.GetInteger(pname))}
}

func (c *Functions) GetBindingi(pname Enum, idx int) Object {
	return Object{uint(c.GetIntegeri(pname, idx))}
}

func (f *Functions) GetError() Enum {
	return Enum(f.glGetError())
}

func (f *Functions) GetRenderbufferParameteri(target, pname Enum) int {
	f.glGetRenderbufferParameteriv(uint32(target), uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetFramebufferAttachmentParameteri(target, attachment, pname Enum) int {
	f.glGetFramebufferAttachmentParameteriv(uint32(target), uint32(attachment), uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetFloat4(pname Enum) [4]float32 {
	f.glGetFloatv(uint32(pname), unsafe.Pointer(&f.floats[0]))
	var r [4]float32
	for i := range r {
		r[i] = f.floats[i]
	}
	return r
}

func (f *Functions) GetFloat(pname Enum) float32 {
	f.glGetFloatv(uint32(pname), unsafe.Pointer(&f.floats[0]))
	return f.floats[0]
}

func (f *Functions) GetInteger4(pname Enum) [4]int {
	f.glGetIntegerv(uint32(pname), unsafe.Pointer(&f.ints[0]))
	var r [4]int
	for i := range r {
		r[i] = int(f.ints[i])
	}
	return r
}

func (f *Functions) GetInteger(pname Enum) int {
	f.glGetIntegerv(uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetIntegeri(pname Enum, idx int) int {
	f.glGetIntegeri_v(uint32(pname), uint32(idx), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetProgrami(p Program, pname Enum) int {
	f.glGetProgramiv(uint32(p.V), uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetProgramBinary(p Program) []byte {
	sz := f.GetProgrami(p, PROGRAM_BINARY_LENGTH)
	if sz == 0 {
		return nil
	}
	buf := make([]byte, sz)
	var format uint32
	f.glGetProgramBinary(uint32(p.V), int32(sz), nil, unsafe.Pointer(&format), unsafe.Pointer(&buf[0]))
	return buf
}

func (f *Functions) GetProgramInfoLog(p Program) string {
	n := f.GetProgrami(p, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	f.glGetProgramInfoLog(uint32(p.V), int32(len(buf)), nil, unsafe.Pointer(&buf[0]))
	return string(buf)
}

func (f *Functions) GetQueryObjectuiv(query Query, pname Enum) uint {
	f.glGetQueryObjectuiv(uint32(query.V), uint32(pname), unsafe.Pointer(&f.uints[0]))
	return uint(f.uints[0])
}

func (f *Functions) GetShaderi(s Shader, pname Enum) int {
	f.glGetShaderiv(uint32(s.V), uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetShaderInfoLog(s Shader) string {
	n := f.GetShaderi(s, INFO_LOG_LENGTH)
	buf := make([]byte, n)
	f.glGetShaderInfoLog(uint32(s.V), int32(len(buf)), nil, unsafe.Pointer(&buf[0]))
	return string(buf)
}

func (f *Functions) getStringi(pname Enum, index int) string {
	return goString(f.glGetStringi(uint32(pname), uint32(index)))
}

func (f *Functions) GetString(pname Enum) string {
	switch {
	case runtime.GOOS == "darwin" && pname == EXTENSIONS:
		// macOS OpenGL 3 core profile doesn't support glGetString(GL_EXTENSIONS).
		// Use glGetStringi(GL_EXTENSIONS, <index>).
		var exts []string
		nexts := f.GetInteger(NUM_EXTENSIONS)
		for i := range nexts {
			ext := f.getStringi(EXTENSIONS, i)
			exts = append(exts, ext)
		}
		return strings.Join(exts, " ")
	default:
		return goString(f.glGetString(uint32(pname)))
	}
}

func (f *Functions) GetUniformBlockIndex(p Program, name string) uint {
	b := cstr(name)
	idx := f.glGetUniformBlockIndex(uint32(p.V), unsafe.Pointer(&b[0]))
	runtime.KeepAlive(b)
	return uint(idx)
}

func (f *Functions) GetUniformLocation(p Program, name string) Uniform {
	b := cstr(name)
	loc := f.glGetUniformLocation(uint32(p.V), unsafe.Pointer(&b[0]))
	runtime.KeepAlive(b)
	return Uniform{int(loc)}
}

func (f *Functions) GetAttribLocation(p Program, name string) Attrib {
	b := cstr(name)
	loc := f.glGetAttribLocation(uint32(p.V), unsafe.Pointer(&b[0]))
	runtime.KeepAlive(b)
	return Attrib(loc)
}

func (f *Functions) GetVertexAttrib(index int, pname Enum) int {
	f.glGetVertexAttribiv(uint32(index), uint32(pname), unsafe.Pointer(&f.ints[0]))
	return int(f.ints[0])
}

func (f *Functions) GetVertexAttribBinding(index int, pname Enum) Object {
	return Object{uint(f.GetVertexAttrib(index, pname))}
}

func (f *Functions) GetVertexAttribPointer(index int, pname Enum) uintptr {
	var ptr uintptr
	f.glGetVertexAttribPointerv(uint32(index), uint32(pname), unsafe.Pointer(&ptr))
	return ptr
}

func (f *Functions) InvalidateFramebuffer(target, attachment Enum) {
	// Framebuffer invalidation is just a hint and can safely be ignored.
	if f.glInvalidateFramebuffer == nil {
		return
	}
	a := uint32(attachment)
	f.glInvalidateFramebuffer(uint32(target), 1, unsafe.Pointer(&a))
}

func (f *Functions) IsEnabled(cap Enum) bool {
	return f.glIsEnabled(uint32(cap)) == TRUE
}

func (f *Functions) LinkProgram(p Program) {
	f.glLinkProgram(uint32(p.V))
}

func (f *Functions) PixelStorei(pname Enum, param int) {
	f.glPixelStorei(uint32(pname), int32(param))
}

func (f *Functions) MemoryBarrier(barriers Enum) {
	f.glMemoryBarrier(uint32(barriers))
}

func (f *Functions) MapBufferRange(target Enum, offset, length int, access Enum) []byte {
	p := f.glMapBufferRange(uint32(target), offset, length, uint32(access))
	if p == nil {
		return nil
	}
	return (*[1 << 30]byte)(p)[:length:length]
}

func (f *Functions) Scissor(x, y, width, height int32) {
	f.glScissor(x, y, width, height)
}

func (f *Functions) ReadPixels(x, y, width, height int, format, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	f.glReadPixels(int32(x), int32(y), int32(width), int32(height), uint32(format), uint32(ty), p)
}

func (f *Functions) RenderbufferStorage(target, internalformat Enum, width, height int) {
	f.glRenderbufferStorage(uint32(target), uint32(internalformat), int32(width), int32(height))
}

func (f *Functions) ShaderSource(s Shader, src string) {
	b := cstr(src)
	cstrp := &b[0]
	strlen := int32(len(src))
	f.glShaderSource(uint32(s.V), 1, unsafe.Pointer(&cstrp), unsafe.Pointer(&strlen))
	runtime.KeepAlive(b)
}

func (f *Functions) TexImage2D(target Enum, level int, internalFormat Enum, width int, height int, format Enum, ty Enum, data []byte) {
	f.glTexImage2D(uint32(target), int32(level), int32(internalFormat), int32(width), int32(height), 0, uint32(format), uint32(ty), unsafe.Pointer(&data[0]))
}

func (f *Functions) TexStorage2D(target Enum, levels int, internalFormat Enum, width, height int) {
	f.glTexStorage2D(uint32(target), int32(levels), uint32(internalFormat), int32(width), int32(height))
}

func (f *Functions) TexSubImage2D(target Enum, level int, x int, y int, width int, height int, format Enum, ty Enum, data []byte) {
	var p unsafe.Pointer
	if len(data) > 0 {
		p = unsafe.Pointer(&data[0])
	}
	f.glTexSubImage2D(uint32(target), int32(level), int32(x), int32(y), int32(width), int32(height), uint32(format), uint32(ty), p)
}

func (f *Functions) TexParameteri(target, pname Enum, param int) {
	f.glTexParameteri(uint32(target), uint32(pname), int32(param))
}

func (f *Functions) UniformBlockBinding(p Program, uniformBlockIndex uint, uniformBlockBinding uint) {
	f.glUniformBlockBinding(uint32(p.V), uint32(uniformBlockIndex), uint32(uniformBlockBinding))
}

func (f *Functions) Uniform1f(dst Uniform, v float32) {
	f.glUniform1f(int32(dst.V), v)
}

func (f *Functions) Uniform1i(dst Uniform, v int) {
	f.glUniform1i(int32(dst.V), int32(v))
}

func (f *Functions) Uniform2f(dst Uniform, v0 float32, v1 float32) {
	f.glUniform2f(int32(dst.V), v0, v1)
}

func (f *Functions) Uniform3f(dst Uniform, v0 float32, v1 float32, v2 float32) {
	f.glUniform3f(int32(dst.V), v0, v1, v2)
}

func (f *Functions) Uniform4f(dst Uniform, v0 float32, v1 float32, v2 float32, v3 float32) {
	f.glUniform4f(int32(dst.V), v0, v1, v2, v3)
}

func (f *Functions) UseProgram(p Program) {
	f.glUseProgram(uint32(p.V))
}

func (f *Functions) UnmapBuffer(target Enum) bool {
	return f.glUnmapBuffer(uint32(target)) == TRUE
}

func (f *Functions) VertexAttribPointer(dst Attrib, size int, ty Enum, normalized bool, stride int, offset int) {
	n := uint8(FALSE)
	if normalized {
		n = uint8(TRUE)
	}
	f.glVertexAttribPointer(uint32(dst), int32(size), uint32(ty), n, int32(stride), uintptr(offset))
}

func (f *Functions) Viewport(x int, y int, width int, height int) {
	f.glViewport(int32(x), int32(y), int32(width), int32(height))
}

// cstr returns a NUL-terminated copy of s suitable for passing to C as a
// const char*.
func cstr(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}
