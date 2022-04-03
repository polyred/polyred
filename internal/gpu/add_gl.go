// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package gpu

/*
#cgo linux pkg-config: x11

#include <stdlib.h>
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
*/
import "C"
import (
	_ "embed"
	"errors"
	"runtime"

	"fmt"

	"poly.red/internal/driver/egl"
	"poly.red/internal/driver/gles"
	"poly.red/math"
)

var device *gles.Functions

func init() {
	ctx = try(newX11EGLContext())
	device = try(gles.NewFunctions())
	addFn = try(newAddShader())
}

// add is a GPU version of math.Mat[float32].Add method.
func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	device.MemoryBarrier(gles.ALL_BARRIER_BITS)
	device.Flush()
	device.Finish()
	return math.Mat[T]{}
}

type shaderFn struct {
	programId gles.Program
}

var (
	// ctx holds a X11 EGL context. It is necessary to initialize EGL context
	// before using OpenGL ES functions.
	ctx *x11Context

	//go:embed add.comp.glsl
	addGl string
	addFn Function
)

func newAddShader() (sf Function, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = ctx.Lock(); err != nil {
		return
	}
	defer ctx.Unlock()

	sid := device.CreateShader(gles.COMPUTE_SHADER)
	if sid.V == 0 {
		err = errors.New("failed to create compute shader")
		return
	}

	device.ShaderSource(sid, addGl)
	device.CompileShader(sid)

	status := device.GetShaderi(sid, gles.COMPILE_STATUS)
	if status == gles.FALSE {
		err = fmt.Errorf("failed to compile shader %v: %v", addGl, device.GetShaderInfoLog(sid))
		return
	}

	pid := device.CreateProgram()
	device.AttachShader(pid, sid)
	device.LinkProgram(pid)
	status = device.GetProgrami(pid, gles.LINK_STATUS)
	if status == gles.FALSE {
		err = fmt.Errorf("failed to link program: %v", device.GetProgramInfoLog(pid))
		return
	}

	device.DeleteShader(sid)
	sf = Function{shaderFn{programId: pid}}
	return
}

type x11Context struct {
	display *C.Display
	eglCtx  *egl.Context
}

func newX11EGLContext() (*x11Context, error) {
	ctx := &x11Context{}

	ctx.display = C.XOpenDisplay(nil)
	if ctx.display == nil {
		panic("x11: cannot connect to the X server")
	}

	var err error
	ctx.eglCtx, err = egl.NewContext(egl.NativeDisplayType(ctx.display))
	if err != nil {
		return nil, err
	}
	return ctx, nil
}

func (c *x11Context) Lock() error {
	return c.eglCtx.MakeCurrent()
}

func (c *x11Context) Unlock() {
	c.eglCtx.ReleaseCurrent()
}
