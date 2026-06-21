// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package gpu

import (
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/ebitengine/purego"
	"poly.red/gpu/shader"
)

// GL constants for the compute path.
const (
	glComputeShader       = 0x91B9
	glShaderStorageBuffer = 0x90D2
	glDynamicRead         = 0x88E9
	glCompileStatus       = 0x8B81
	glLinkStatus          = 0x8B82
	glInfoLogLength       = 0x8B84
	glMapReadBit          = 0x0001
	glAllBarrierBits      = 0xFFFFFFFF
)

// TestGLComputeEndToEnd proves the full cgo-free GLES compute path on the CI
// runner: it takes a Go kernel, compiles it to GLSL with shader.CompileGLSL,
// then through purego it compiles+links the compute program, uploads the input
// SSBOs, dispatches one invocation per element, barriers, maps the output buffer
// back, and checks the result against the CPU. This is the discriminator the GL
// backend (item #2) is built on: it validates both the GLSL emitter against a
// real driver and the buffer/dispatch/readback mechanics.
func TestGLComputeEndToEnd(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the headless GLES compute test")
	}
	// eglMakeCurrent binds the context to this OS thread; buffer marshaling below
	// allocates and could migrate the goroutine, leaving GL calls without a
	// current context. Pin the thread for the whole test.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	egl, err := purego.Dlopen("libEGL.so.1", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libEGL.so.1 not available: %v", err)
	}
	gles, err := purego.Dlopen("libGLESv2.so.2", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Skipf("libGLESv2.so.2 not available: %v", err)
	}
	sym := func(h uintptr, name string) uintptr {
		p, e := purego.Dlsym(h, name)
		if e != nil {
			t.Fatalf("dlsym %s: %v", name, e)
		}
		return p
	}

	// --- EGL surfaceless context (cgo-free) ---
	eglGetDisplay := sym(egl, "eglGetDisplay")
	eglInitialize := sym(egl, "eglInitialize")
	eglBindAPI := sym(egl, "eglBindAPI")
	eglChooseConfig := sym(egl, "eglChooseConfig")
	eglCreateContext := sym(egl, "eglCreateContext")
	eglMakeCurrent := sym(egl, "eglMakeCurrent")

	dpy, _, _ := purego.SyscallN(eglGetDisplay, uintptr(eglDefaultDisplay))
	if dpy == 0 {
		t.Fatal("eglGetDisplay returned EGL_NO_DISPLAY")
	}
	var major, minor int32
	if r, _, _ := purego.SyscallN(eglInitialize, dpy, uintptr(unsafe.Pointer(&major)), uintptr(unsafe.Pointer(&minor))); r == 0 {
		t.Fatal("eglInitialize failed")
	}
	purego.SyscallN(eglBindAPI, uintptr(eglOpenGLESAPI))
	cfgAttribs := []int32{eglRenderableType, eglOpenGLES3Bit, eglSurfaceType, eglPbufferBit, eglRedSize, 8, eglGreenSize, 8, eglBlueSize, 8, eglNone}
	var cfg uintptr
	var n int32
	if r, _, _ := purego.SyscallN(eglChooseConfig, dpy, uintptr(unsafe.Pointer(&cfgAttribs[0])), uintptr(unsafe.Pointer(&cfg)), 1, uintptr(unsafe.Pointer(&n))); r == 0 || n == 0 {
		t.Fatal("eglChooseConfig failed")
	}
	ctxAttribs := []int32{eglContextMajor, 3, eglNone}
	ctx, _, _ := purego.SyscallN(eglCreateContext, dpy, cfg, uintptr(eglNoContext), uintptr(unsafe.Pointer(&ctxAttribs[0])))
	if ctx == 0 {
		t.Fatal("eglCreateContext failed")
	}
	if r, _, _ := purego.SyscallN(eglMakeCurrent, dpy, uintptr(eglNoSurface), uintptr(eglNoSurface), ctx); r == 0 {
		t.Fatal("eglMakeCurrent failed")
	}

	// --- GLES entry points ---
	glCreateShader := sym(gles, "glCreateShader")
	glShaderSource := sym(gles, "glShaderSource")
	glCompileShader := sym(gles, "glCompileShader")
	glGetShaderiv := sym(gles, "glGetShaderiv")
	glGetShaderInfoLog := sym(gles, "glGetShaderInfoLog")
	glCreateProgram := sym(gles, "glCreateProgram")
	glAttachShader := sym(gles, "glAttachShader")
	glLinkProgram := sym(gles, "glLinkProgram")
	glGetProgramiv := sym(gles, "glGetProgramiv")
	glUseProgram := sym(gles, "glUseProgram")
	glGenBuffers := sym(gles, "glGenBuffers")
	glBindBuffer := sym(gles, "glBindBuffer")
	glBufferData := sym(gles, "glBufferData")
	glBindBufferBase := sym(gles, "glBindBufferBase")
	glDispatchCompute := sym(gles, "glDispatchCompute")
	glMemoryBarrier := sym(gles, "glMemoryBarrier")
	glMapBufferRange := sym(gles, "glMapBufferRange")
	glUnmapBuffer := sym(gles, "glUnmapBuffer")

	// --- Go kernel -> GLSL ---
	ks, err := shader.CompileGLSL(`package kernels
func Add(gid uint, a []float32, b []float32, out []float32) {
	out[gid] = a[gid] + b[gid]
}`)
	if err != nil {
		t.Fatalf("CompileGLSL: %v", err)
	}
	k := ks["Add"]
	glsl := k.GLSL

	// --- compile + link the compute program ---
	sh, _, _ := purego.SyscallN(glCreateShader, uintptr(glComputeShader))
	psrc := &glsl
	slen := int32(len(glsl))
	purego.SyscallN(glShaderSource, sh, 1, uintptr(unsafe.Pointer(psrc)), uintptr(unsafe.Pointer(&slen)))
	runtime.KeepAlive(psrc)
	purego.SyscallN(glCompileShader, sh)
	var status int32
	purego.SyscallN(glGetShaderiv, sh, glCompileStatus, uintptr(unsafe.Pointer(&status)))
	if status == 0 {
		var logLen int32
		purego.SyscallN(glGetShaderiv, sh, glInfoLogLength, uintptr(unsafe.Pointer(&logLen)))
		msg := make([]byte, logLen+1)
		if logLen > 0 {
			purego.SyscallN(glGetShaderInfoLog, sh, uintptr(logLen), 0, uintptr(unsafe.Pointer(&msg[0])))
		}
		t.Fatalf("compute shader compile failed: %s\n--- GLSL ---\n%s", cStr(uintptr(unsafe.Pointer(&msg[0]))), glsl)
	}
	prog, _, _ := purego.SyscallN(glCreateProgram)
	purego.SyscallN(glAttachShader, prog, sh)
	purego.SyscallN(glLinkProgram, prog)
	purego.SyscallN(glGetProgramiv, prog, glLinkStatus, uintptr(unsafe.Pointer(&status)))
	if status == 0 {
		t.Fatal("compute program link failed")
	}
	purego.SyscallN(glUseProgram, prog)

	// --- input data ---
	const count = 256
	a := make([]float32, count)
	b := make([]float32, count)
	for i := range a {
		a[i] = float32(i)
		b[i] = float32(2*i + 1)
	}
	out := make([]float32, count)

	// Map each kernel binding (by name) to its uploaded SSBO.
	data := map[string][]float32{"a": a, "b": b, "out": out}
	mkSSBO := func(d []float32) uint32 {
		var buf uint32
		purego.SyscallN(glGenBuffers, 1, uintptr(unsafe.Pointer(&buf)))
		purego.SyscallN(glBindBuffer, uintptr(glShaderStorageBuffer), uintptr(buf))
		purego.SyscallN(glBufferData, uintptr(glShaderStorageBuffer), uintptr(len(d)*4), uintptr(unsafe.Pointer(&d[0])), uintptr(glDynamicRead))
		runtime.KeepAlive(d)
		return buf
	}
	bufs := map[string]uint32{}
	for _, bd := range k.Bindings {
		buf := mkSSBO(data[bd.Name])
		bufs[bd.Name] = buf
		purego.SyscallN(glBindBufferBase, uintptr(glShaderStorageBuffer), uintptr(bd.Index), uintptr(buf))
	}

	// --- dispatch one invocation per element, barrier ---
	purego.SyscallN(glDispatchCompute, count, 1, 1)
	purego.SyscallN(glMemoryBarrier, uintptr(uint32(glAllBarrierBits)))

	// --- read back the output SSBO ---
	purego.SyscallN(glBindBuffer, uintptr(glShaderStorageBuffer), uintptr(bufs["out"]))
	p, _, _ := purego.SyscallN(glMapBufferRange, uintptr(glShaderStorageBuffer), 0, uintptr(count*4), uintptr(glMapReadBit))
	if p == 0 {
		t.Fatal("glMapBufferRange returned nil")
	}
	mapped := unsafe.Slice((*float32)(unsafe.Pointer(p)), count)
	copy(out, mapped)
	purego.SyscallN(glUnmapBuffer, uintptr(glShaderStorageBuffer))

	// --- verify against the CPU ---
	bad := 0
	for i := range out {
		want := a[i] + b[i]
		if out[i] != want {
			if bad < 5 {
				t.Errorf("out[%d] = %v, want %v", i, out[i], want)
			}
			bad++
		}
	}
	if bad == 0 {
		t.Logf("GLES compute add: %d/%d elements match the CPU", count, count)
	}
}
