// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// This file is the cgo-free OpenGL ES 3.1 compute backend. It reaches EGL/GLES
// through purego (no cgo), creates a headless surfaceless context (works on a
// software rasterizer such as Mesa llvmpipe, which is how it is verified in CI),
// and implements the private backend interface for the compute pipeline. GL is
// thread-bound, so the context lives on one locked OS thread and every GL call
// is marshaled onto it via do(). See specs/foundations/gpu-gl-backend.md.
package gpu

import (
	"fmt"
	"runtime"
	"unsafe"

	"github.com/ebitengine/purego"
)

// EGL / GLES constants.
const (
	eglDefaultDisplay = 0
	eglNoContext      = 0
	eglNoSurface      = 0
	eglOpenGLESAPI    = 0x30A0
	eglNone           = 0x3038
	eglContextMajor   = 0x3098
	eglRenderableType = 0x3040
	eglOpenGLES3Bit   = 0x0040
	eglSurfaceType    = 0x3033
	eglPbufferBit     = 0x0001
	eglRedSize        = 0x3024
	eglGreenSize      = 0x3023
	eglBlueSize       = 0x3022

	glComputeShader                  = 0x91B9
	glShaderStorageBuffer            = 0x90D2
	glUniformBuffer                  = 0x8A11
	glDynamicRead                    = 0x88E9
	glCompileStatus                  = 0x8B81
	glLinkStatus                     = 0x8B82
	glInfoLogLength                  = 0x8B84
	glMapReadBit                     = 0x0001
	glAllBarrierBits                 = 0xFFFFFFFF
	glMaxComputeWorkGroupInvocations = 0x90EB

	glFramebuffer       = 0x8D40
	glColorAttachment0  = 0x8CE0
	glTexture2D         = 0x0DE1
	glRGBA              = 0x1908
	glRGBA8             = 0x8058
	glUnsignedByte      = 0x1401
	glNearest           = 0x2600
	glTexMinFilter      = 0x2801
	glTexMagFilter      = 0x2800
	glPoints            = 0x0000
	glLines             = 0x0001
	glTriangles         = 0x0004
	glTriangleStripEnum = 0x0005
	glColor             = 0x1800 // GL_COLOR, for glClearBufferfv
)

// glFns holds the resolved EGL/GLES entry points (purego function pointers).
type glFns struct {
	eglGetDisplay, eglInitialize, eglBindAPI, eglChooseConfig    uintptr
	eglCreateContext, eglMakeCurrent, eglDestroyContext, eglTerm uintptr

	createShader, shaderSource, compileShader, getShaderiv, getShaderInfoLog uintptr
	createProgram, attachShader, linkProgram, getProgramiv, useProgram       uintptr
	deleteShader, deleteProgram                                              uintptr
	genBuffers, deleteBuffers, bindBuffer, bufferData, bindBufferBase        uintptr
	dispatchCompute, memoryBarrier, mapBufferRange, unmapBuffer              uintptr
	finish, getIntegerv                                                      uintptr

	genTextures, bindTexture, texImage2D, texParameteri                      uintptr
	genFramebuffers, bindFramebuffer, framebufferTexture2D, checkFramebuffer uintptr
	genVertexArrays, bindVertexArray                                         uintptr
	viewport, clearBufferfv, drawArrays, readPixels                          uintptr
}

type glBackend struct {
	reqs chan func()
	fns  glFns
	dpy  uintptr
	ctx  uintptr
}

func openBackend(d Driver) (backend, Driver, error) {
	if d == DriverVulkan {
		vb, err := newVKBackend()
		if err != nil {
			return nil, DriverAuto, err
		}
		return vb, DriverVulkan, nil
	}
	if d != DriverAuto && d != DriverGL {
		return nil, DriverAuto, ErrUnsupported
	}
	b := &glBackend{reqs: make(chan func())}
	ready := make(chan error, 1)
	go b.loop(ready)
	if err := <-ready; err != nil {
		return nil, DriverAuto, err
	}
	return b, DriverGL, nil
}

// loop owns the GL context on a single locked OS thread and serially runs the
// closures submitted by do(). eglMakeCurrent binds the context to this thread;
// keeping all GL work here is what makes the backend safe to call from any
// goroutine.
func (b *glBackend) loop(ready chan error) {
	runtime.LockOSThread()
	if err := b.init(); err != nil {
		ready <- err
		return
	}
	ready <- nil
	for fn := range b.reqs {
		fn()
	}
}

func (b *glBackend) do(fn func()) {
	done := make(chan struct{})
	b.reqs <- func() { defer close(done); fn() }
	<-done
}

func (b *glBackend) init() error {
	egl, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("gpu/gl: %w", err)
	}
	gles, err := purego.Dlopen("libGLESv2.so.2", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return fmt.Errorf("gpu/gl: %w", err)
	}
	var loadErr error
	sym := func(h uintptr, name string) uintptr {
		p, e := purego.Dlsym(h, name)
		if e != nil && loadErr == nil {
			loadErr = fmt.Errorf("gpu/gl: dlsym %s: %w", name, e)
		}
		return p
	}
	f := &b.fns
	f.eglGetDisplay = sym(egl, "eglGetDisplay")
	f.eglInitialize = sym(egl, "eglInitialize")
	f.eglBindAPI = sym(egl, "eglBindAPI")
	f.eglChooseConfig = sym(egl, "eglChooseConfig")
	f.eglCreateContext = sym(egl, "eglCreateContext")
	f.eglMakeCurrent = sym(egl, "eglMakeCurrent")
	f.eglDestroyContext = sym(egl, "eglDestroyContext")
	f.eglTerm = sym(egl, "eglTerminate")
	f.createShader = sym(gles, "glCreateShader")
	f.shaderSource = sym(gles, "glShaderSource")
	f.compileShader = sym(gles, "glCompileShader")
	f.getShaderiv = sym(gles, "glGetShaderiv")
	f.getShaderInfoLog = sym(gles, "glGetShaderInfoLog")
	f.createProgram = sym(gles, "glCreateProgram")
	f.attachShader = sym(gles, "glAttachShader")
	f.linkProgram = sym(gles, "glLinkProgram")
	f.getProgramiv = sym(gles, "glGetProgramiv")
	f.useProgram = sym(gles, "glUseProgram")
	f.deleteShader = sym(gles, "glDeleteShader")
	f.deleteProgram = sym(gles, "glDeleteProgram")
	f.genBuffers = sym(gles, "glGenBuffers")
	f.deleteBuffers = sym(gles, "glDeleteBuffers")
	f.bindBuffer = sym(gles, "glBindBuffer")
	f.bufferData = sym(gles, "glBufferData")
	f.bindBufferBase = sym(gles, "glBindBufferBase")
	f.dispatchCompute = sym(gles, "glDispatchCompute")
	f.memoryBarrier = sym(gles, "glMemoryBarrier")
	f.mapBufferRange = sym(gles, "glMapBufferRange")
	f.unmapBuffer = sym(gles, "glUnmapBuffer")
	f.finish = sym(gles, "glFinish")
	f.getIntegerv = sym(gles, "glGetIntegerv")
	f.genTextures = sym(gles, "glGenTextures")
	f.bindTexture = sym(gles, "glBindTexture")
	f.texImage2D = sym(gles, "glTexImage2D")
	f.texParameteri = sym(gles, "glTexParameteri")
	f.genFramebuffers = sym(gles, "glGenFramebuffers")
	f.bindFramebuffer = sym(gles, "glBindFramebuffer")
	f.framebufferTexture2D = sym(gles, "glFramebufferTexture2D")
	f.checkFramebuffer = sym(gles, "glCheckFramebufferStatus")
	f.genVertexArrays = sym(gles, "glGenVertexArrays")
	f.bindVertexArray = sym(gles, "glBindVertexArray")
	f.viewport = sym(gles, "glViewport")
	f.clearBufferfv = sym(gles, "glClearBufferfv")
	f.drawArrays = sym(gles, "glDrawArrays")
	f.readPixels = sym(gles, "glReadPixels")
	if loadErr != nil {
		return loadErr
	}

	dpy, _, _ := purego.SyscallN(f.eglGetDisplay, uintptr(eglDefaultDisplay))
	if dpy == 0 {
		return fmt.Errorf("gpu/gl: eglGetDisplay returned EGL_NO_DISPLAY (need EGL_PLATFORM=surfaceless or a display)")
	}
	var major, minor int32
	if r, _, _ := purego.SyscallN(f.eglInitialize, dpy, uintptr(unsafe.Pointer(&major)), uintptr(unsafe.Pointer(&minor))); r == 0 {
		return fmt.Errorf("gpu/gl: eglInitialize failed")
	}
	purego.SyscallN(f.eglBindAPI, uintptr(eglOpenGLESAPI))
	cfgAttribs := []int32{eglRenderableType, eglOpenGLES3Bit, eglSurfaceType, eglPbufferBit, eglRedSize, 8, eglGreenSize, 8, eglBlueSize, 8, eglNone}
	var cfg uintptr
	var n int32
	if r, _, _ := purego.SyscallN(f.eglChooseConfig, dpy, uintptr(unsafe.Pointer(&cfgAttribs[0])), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&n))); r == 0 || n == 0 {
		return fmt.Errorf("gpu/gl: eglChooseConfig found no config")
	}
	ctxAttribs := []int32{eglContextMajor, 3, eglNone}
	ctx, _, _ := purego.SyscallN(f.eglCreateContext, dpy, cfg, uintptr(eglNoContext), uintptr(unsafe.Pointer(&ctxAttribs[0])))
	if ctx == 0 {
		return fmt.Errorf("gpu/gl: eglCreateContext failed")
	}
	if r, _, _ := purego.SyscallN(f.eglMakeCurrent, dpy, uintptr(eglNoSurface), uintptr(eglNoSurface), ctx); r == 0 {
		return fmt.Errorf("gpu/gl: eglMakeCurrent failed")
	}
	// GLES requires a bound vertex array object for draw calls; the vertex data
	// itself comes from a storage buffer indexed by gl_VertexID (matching the
	// Metal model), so this VAO carries no attributes.
	var vao uint32
	purego.SyscallN(f.genVertexArrays, 1, uintptr(unsafe.Pointer(&vao)))
	purego.SyscallN(f.bindVertexArray, uintptr(vao))
	b.dpy, b.ctx = dpy, ctx
	return nil
}

// --- backend interface ---

type glBuffer struct {
	b      *glBackend
	id     uint32
	size   int
	target uintptr // GL_SHADER_STORAGE_BUFFER or GL_UNIFORM_BUFFER
}

func (b *glBackend) newBuffer(size int, usage BufferUsage, data []byte) (backendBuffer, error) {
	// SSBO vs UBO follows the buffer's usage. GL keeps these in separate binding
	// namespaces, so the buffer carries its target and the command buffer binds
	// each one accordingly (see commit).
	target := uintptr(glShaderStorageBuffer)
	if usage&BufferUniform != 0 {
		target = uintptr(glUniformBuffer)
	}
	buf := &glBuffer{b: b, size: size, target: target}
	b.do(func() {
		f := &b.fns
		purego.SyscallN(f.genBuffers, 1, uintptr(unsafe.Pointer(&buf.id)))
		purego.SyscallN(f.bindBuffer, buf.target, uintptr(buf.id))
		var p unsafe.Pointer
		if len(data) > 0 {
			p = unsafe.Pointer(&data[0])
		}
		purego.SyscallN(f.bufferData, buf.target, uintptr(size), uintptr(p), uintptr(glDynamicRead))
		runtime.KeepAlive(data)
	})
	return buf, nil
}

func (b *glBuffer) bytes() []byte {
	out := make([]byte, b.size)
	b.b.do(func() {
		f := &b.b.fns
		purego.SyscallN(f.bindBuffer, b.target, uintptr(b.id))
		p, _, _ := purego.SyscallN(f.mapBufferRange, b.target, 0, uintptr(b.size), uintptr(glMapReadBit))
		if p != 0 {
			copy(out, unsafe.Slice((*byte)(unsafe.Pointer(p)), b.size))
			purego.SyscallN(f.unmapBuffer, b.target)
		}
	})
	return out
}

func (b *glBuffer) release() {
	b.b.do(func() {
		purego.SyscallN(b.b.fns.deleteBuffers, 1, uintptr(unsafe.Pointer(&b.id)))
	})
}

type glShaderModule struct{ glsl string }

func (glShaderModule) isShaderModule() {}

func (b *glBackend) newShaderModule(src ShaderSource) (backendShaderModule, error) {
	if src.GLSL == "" {
		return nil, fmt.Errorf("gpu/gl: ShaderSource.GLSL is empty (the GL backend needs GLSL; use shader.CompileGLSL)")
	}
	return glShaderModule{glsl: src.GLSL}, nil
}

type glComputePipeline struct {
	program uint32
	maxThr  int
}

func (p glComputePipeline) maxThreads() int { return p.maxThr }

func (b *glBackend) newComputePipeline(mod backendShaderModule, entry string) (backendComputePipeline, error) {
	gm, ok := mod.(glShaderModule)
	if !ok {
		return nil, fmt.Errorf("gpu/gl: shader module is not a GL module")
	}
	var prog uint32
	var compileErr error
	var maxThr int32
	b.do(func() {
		f := &b.fns
		sh, _, _ := purego.SyscallN(f.createShader, uintptr(glComputeShader))
		src := gm.glsl
		psrc := &src
		slen := int32(len(src))
		purego.SyscallN(f.shaderSource, sh, 1, uintptr(unsafe.Pointer(psrc)), uintptr(unsafe.Pointer(&slen)))
		runtime.KeepAlive(psrc)
		purego.SyscallN(f.compileShader, sh)
		var status int32
		purego.SyscallN(f.getShaderiv, sh, glCompileStatus, uintptr(unsafe.Pointer(&status)))
		if status == 0 {
			compileErr = fmt.Errorf("gpu/gl: compute shader compile failed: %s", b.shaderLog(sh))
			purego.SyscallN(f.deleteShader, sh)
			return
		}
		p, _, _ := purego.SyscallN(f.createProgram)
		purego.SyscallN(f.attachShader, p, sh)
		purego.SyscallN(f.linkProgram, p)
		purego.SyscallN(f.getProgramiv, p, glLinkStatus, uintptr(unsafe.Pointer(&status)))
		purego.SyscallN(f.deleteShader, sh)
		if status == 0 {
			compileErr = fmt.Errorf("gpu/gl: compute program link failed")
			purego.SyscallN(f.deleteProgram, p)
			return
		}
		prog = uint32(p)
		purego.SyscallN(f.getIntegerv, uintptr(glMaxComputeWorkGroupInvocations), uintptr(unsafe.Pointer(&maxThr)))
	})
	if compileErr != nil {
		return nil, compileErr
	}
	return glComputePipeline{program: prog, maxThr: int(maxThr)}, nil
}

// shaderLog reads a shader's info log; must run on the context thread.
func (b *glBackend) shaderLog(sh uintptr) string {
	f := &b.fns
	var logLen int32
	purego.SyscallN(f.getShaderiv, sh, glInfoLogLength, uintptr(unsafe.Pointer(&logLen)))
	if logLen <= 0 {
		return "(no log)"
	}
	msg := make([]byte, logLen)
	purego.SyscallN(f.getShaderInfoLog, sh, uintptr(logLen), 0, uintptr(unsafe.Pointer(&msg[0])))
	return cStr(uintptr(unsafe.Pointer(&msg[0])))
}

func (b *glBackend) waitIdle() {
	b.do(func() { purego.SyscallN(b.fns.finish) })
}

func (b *glBackend) close() error {
	b.do(func() {
		f := &b.fns
		purego.SyscallN(f.eglMakeCurrent, b.dpy, uintptr(eglNoSurface), uintptr(eglNoSurface), uintptr(eglNoContext))
		purego.SyscallN(f.eglDestroyContext, b.dpy, b.ctx)
		purego.SyscallN(f.eglTerm, b.dpy)
	})
	close(b.reqs)
	return nil
}

// --- command buffer (record then replay on commit) ---

// glCmd records GL operations as closures and replays them on the context thread
// at commit. This serves both compute passes and render passes uniformly.
type glCmd struct {
	b    *glBackend
	ops  []func()
	prog uint32 // current compute pipeline program
	gx   int    // current dispatch x
}

func (b *glBackend) newCommandBuffer() backendCommandBuffer { return &glCmd{b: b} }

func (c *glCmd) record(fn func()) { c.ops = append(c.ops, fn) }

func (c *glCmd) commit() {
	c.b.do(func() {
		for _, op := range c.ops {
			op()
		}
		purego.SyscallN(c.b.fns.finish)
	})
}

// --- compute pass ---

func (c *glCmd) beginCompute() {}

func (c *glCmd) setComputePipeline(p backendComputePipeline) {
	prog := p.(glComputePipeline).program
	c.record(func() { purego.SyscallN(c.b.fns.useProgram, uintptr(prog)) })
}

func (c *glCmd) setBuffer(buf backendBuffer, offset, index int) {
	gb := buf.(*glBuffer)
	c.record(func() { purego.SyscallN(c.b.fns.bindBufferBase, gb.target, uintptr(index), uintptr(gb.id)) })
}

func (c *glCmd) dispatch(x, y, z int) {
	c.record(func() {
		f := &c.b.fns
		purego.SyscallN(f.dispatchCompute, uintptr(x), 1, 1)
		purego.SyscallN(f.memoryBarrier, uintptr(uint32(glAllBarrierBits)))
	})
}

func (c *glCmd) endCompute() {}

// --- render support ---

func glPrim(p Primitive) uintptr {
	switch p {
	case TriangleStrip:
		return glTriangleStripEnum
	case LineList:
		return glLines
	case PointList:
		return glPoints
	default:
		return glTriangles
	}
}

type glTexture struct {
	b    *glBackend
	id   uint32
	fbo  uint32
	w, h int
}

func (b *glBackend) newTexture(format TextureFormat, w, h int, renderTarget bool) (backendTexture, error) {
	t := &glTexture{b: b, w: w, h: h}
	b.do(func() {
		f := &b.fns
		purego.SyscallN(f.genTextures, 1, uintptr(unsafe.Pointer(&t.id)))
		purego.SyscallN(f.bindTexture, uintptr(glTexture2D), uintptr(t.id))
		purego.SyscallN(f.texImage2D, uintptr(glTexture2D), 0, uintptr(glRGBA8), uintptr(w), uintptr(h), 0, uintptr(glRGBA), uintptr(glUnsignedByte), 0)
		purego.SyscallN(f.texParameteri, uintptr(glTexture2D), uintptr(glTexMinFilter), uintptr(glNearest))
		purego.SyscallN(f.texParameteri, uintptr(glTexture2D), uintptr(glTexMagFilter), uintptr(glNearest))
		if renderTarget {
			purego.SyscallN(f.genFramebuffers, 1, uintptr(unsafe.Pointer(&t.fbo)))
			purego.SyscallN(f.bindFramebuffer, uintptr(glFramebuffer), uintptr(t.fbo))
			purego.SyscallN(f.framebufferTexture2D, uintptr(glFramebuffer), uintptr(glColorAttachment0), uintptr(glTexture2D), uintptr(t.id), 0)
		}
	})
	return t, nil
}

func (t *glTexture) readPixels() []byte {
	dst := make([]byte, t.w*t.h*4)
	t.b.do(func() {
		f := &t.b.fns
		purego.SyscallN(f.bindFramebuffer, uintptr(glFramebuffer), uintptr(t.fbo))
		purego.SyscallN(f.readPixels, 0, 0, uintptr(t.w), uintptr(t.h), uintptr(glRGBA), uintptr(glUnsignedByte), uintptr(unsafe.Pointer(&dst[0])))
	})
	// GL's framebuffer origin is bottom-left; flip rows so the result is
	// top-down, matching the Metal backend and image.RGBA.
	row := t.w * 4
	flipped := make([]byte, len(dst))
	for y := 0; y < t.h; y++ {
		copy(flipped[y*row:(y+1)*row], dst[(t.h-1-y)*row:(t.h-y)*row])
	}
	return flipped
}

func (t *glTexture) write(pixels []byte, bytesPerRow int) {
	t.b.do(func() {
		f := &t.b.fns
		purego.SyscallN(f.bindTexture, uintptr(glTexture2D), uintptr(t.id))
		purego.SyscallN(f.texImage2D, uintptr(glTexture2D), 0, uintptr(glRGBA8), uintptr(t.w), uintptr(t.h), 0, uintptr(glRGBA), uintptr(glUnsignedByte), uintptr(unsafe.Pointer(&pixels[0])))
		runtime.KeepAlive(pixels)
	})
}

type glRenderPipeline struct{ program uint32 }

func (glRenderPipeline) isRenderPipeline() {}

func (b *glBackend) newRenderPipeline(vmod backendShaderModule, ventry string, fmod backendShaderModule, fentry string, color TextureFormat, extraColor []TextureFormat, depth TextureFormat) (backendRenderPipeline, error) {
	vs, ok1 := vmod.(glShaderModule)
	fs, ok2 := fmod.(glShaderModule)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("gpu/gl: render pipeline needs GL shader modules")
	}
	var prog uint32
	var perr error
	b.do(func() { prog, perr = b.linkRender(vs.glsl, fs.glsl) })
	if perr != nil {
		return nil, perr
	}
	return glRenderPipeline{program: prog}, nil
}

// linkRender compiles a vertex+fragment program; must run on the context thread.
func (b *glBackend) linkRender(vsrc, fsrc string) (uint32, error) {
	f := &b.fns
	const glVertexShader = 0x8B31
	const glFragmentShader = 0x8B30
	compile := func(kind uintptr, src string) (uintptr, error) {
		sh, _, _ := purego.SyscallN(f.createShader, kind)
		psrc := &src
		slen := int32(len(src))
		purego.SyscallN(f.shaderSource, sh, 1, uintptr(unsafe.Pointer(psrc)), uintptr(unsafe.Pointer(&slen)))
		runtime.KeepAlive(psrc)
		purego.SyscallN(f.compileShader, sh)
		var status int32
		purego.SyscallN(f.getShaderiv, sh, glCompileStatus, uintptr(unsafe.Pointer(&status)))
		if status == 0 {
			return 0, fmt.Errorf("gpu/gl: shader compile failed: %s", b.shaderLog(sh))
		}
		return sh, nil
	}
	vs, err := compile(glVertexShader, vsrc)
	if err != nil {
		return 0, err
	}
	fsh, err := compile(glFragmentShader, fsrc)
	if err != nil {
		return 0, err
	}
	p, _, _ := purego.SyscallN(f.createProgram)
	purego.SyscallN(f.attachShader, p, vs)
	purego.SyscallN(f.attachShader, p, fsh)
	purego.SyscallN(f.linkProgram, p)
	purego.SyscallN(f.deleteShader, vs)
	purego.SyscallN(f.deleteShader, fsh)
	var status int32
	purego.SyscallN(f.getProgramiv, p, glLinkStatus, uintptr(unsafe.Pointer(&status)))
	if status == 0 {
		purego.SyscallN(f.deleteProgram, p)
		return 0, fmt.Errorf("gpu/gl: render program link failed")
	}
	return uint32(p), nil
}

func (b *glBackend) newSampler(desc SamplerDescriptor) backendSampler { return nil }

func (c *glCmd) beginRender(info renderPassInfo) {
	t := info.color.(*glTexture)
	clear := info.load == LoadClear
	cc := info.clearColor
	c.record(func() {
		f := &c.b.fns
		purego.SyscallN(f.bindFramebuffer, uintptr(glFramebuffer), uintptr(t.fbo))
		purego.SyscallN(f.viewport, 0, 0, uintptr(t.w), uintptr(t.h))
		if clear {
			// glClearBufferfv takes a *GLfloat (an integer-register pointer arg),
			// so it is safe through SyscallN, unlike glClearColor's float args.
			vals := [4]float32{float32(cc[0]), float32(cc[1]), float32(cc[2]), float32(cc[3])}
			purego.SyscallN(f.clearBufferfv, uintptr(glColor), 0, uintptr(unsafe.Pointer(&vals[0])))
		}
	})
}

func (c *glCmd) setRenderPipeline(p backendRenderPipeline) {
	prog := p.(glRenderPipeline).program
	c.record(func() { purego.SyscallN(c.b.fns.useProgram, uintptr(prog)) })
}

func (c *glCmd) setRenderBuffer(buf backendBuffer, offset, index int) {
	gb := buf.(*glBuffer)
	c.record(func() { purego.SyscallN(c.b.fns.bindBufferBase, gb.target, uintptr(index), uintptr(gb.id)) })
}

func (c *glCmd) setVertexBuffer(buf backendBuffer, index int) {
	gb := buf.(*glBuffer)
	c.record(func() {
		purego.SyscallN(c.b.fns.bindBufferBase, uintptr(glShaderStorageBuffer), uintptr(index), uintptr(gb.id))
	})
}

func (c *glCmd) draw(prim Primitive, start, count int) {
	mode := glPrim(prim)
	c.record(func() { purego.SyscallN(c.b.fns.drawArrays, mode, uintptr(start), uintptr(count)) })
}

func (c *glCmd) endRender() {}

func (c *glCmd) setComputeTexture(index int, t backendTexture) {}
func (c *glCmd) setComputeSampler(index int, s backendSampler) {}

// cStr converts a NUL-terminated C string at p to a Go string.
func cStr(p uintptr) string {
	if p == 0 {
		return ""
	}
	var b []byte
	for i := 0; ; i++ {
		ch := *(*byte)(unsafe.Pointer(p + uintptr(i)))
		if ch == 0 {
			break
		}
		b = append(b, ch)
	}
	return string(b)
}
