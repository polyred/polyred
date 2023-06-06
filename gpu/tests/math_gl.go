// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package tests

import (
	_ "embed"
	"runtime"

	"poly.red/internal/driver/egl"
	"poly.red/internal/driver/gles"
	"poly.red/math"
)

var device *gles.Functions

func init() {
	ctx = try(egl.NewContext(egl.NewDisplay()))
	device = try(gles.NewFunctions())
	mathLib = try(newMathShader())
}

// add is a GPU version of math.Mat[float32].Add method.
func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.MakeCurrent(); err != nil {
		panic("gpu: cannot establish egl context")
	}
	defer ctx.ReleaseCurrent()

	data := make([]T, len(m1.Data))
	out := newBuffer(gles.SHADER_STORAGE_BUFFER, len(data), gles.STREAM_COPY, data)
	bufA := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m1.Data), gles.STREAM_COPY, m1.Data)
	bufB := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m2.Data), gles.STREAM_COPY, m2.Data)

	device.BindBufferBase(out.target, 0, out.id)
	device.BindBufferBase(bufA.target, 1, bufA.id)
	device.BindBufferBase(bufB.target, 2, bufB.id)

	device.UseProgram(mathLib.funcAdd.progId)
	device.DispatchCompute(65536*8, 1, 1)

	device.MemoryBarrier(gles.ALL_BARRIER_BITS)
	device.Flush()
	device.Finish()
	readGPU(out, data)

	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: data,
	}
}

// sub is a GPU version of math.Mat[float32].Sub method.
func sub[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.MakeCurrent(); err != nil {
		panic("gpu: cannot establish egl context")
	}
	defer ctx.ReleaseCurrent()

	data := make([]T, len(m1.Data))
	out := newBuffer(gles.SHADER_STORAGE_BUFFER, len(data), gles.STREAM_COPY, data)
	bufA := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m1.Data), gles.STREAM_COPY, m1.Data)
	bufB := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m2.Data), gles.STREAM_COPY, m2.Data)

	device.BindBufferBase(out.target, 0, out.id)
	device.BindBufferBase(bufA.target, 1, bufA.id)
	device.BindBufferBase(bufB.target, 2, bufB.id)

	device.UseProgram(mathLib.funcSub.progId)
	device.DispatchCompute(65536*8, 1, 1)

	device.MemoryBarrier(gles.ALL_BARRIER_BITS)
	device.Flush()
	device.Finish()
	readGPU(out, data)

	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: data,
	}
}

// sqrt is a GPU version of math.Mat[float32].Sub method.
func sqrt[T DataType](m math.Mat[T]) math.Mat[T] {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.MakeCurrent(); err != nil {
		panic("gpu: cannot establish egl context")
	}
	defer ctx.ReleaseCurrent()

	data := make([]T, len(m.Data))
	out := newBuffer(gles.SHADER_STORAGE_BUFFER, len(data), gles.STREAM_COPY, data)
	bufA := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m.Data), gles.STREAM_COPY, m.Data)

	device.BindBufferBase(out.target, 0, out.id)
	device.BindBufferBase(bufA.target, 1, bufA.id)

	device.UseProgram(mathLib.funcSqrt.progId)
	device.DispatchCompute(65536*8, 1, 1)

	device.MemoryBarrier(gles.ALL_BARRIER_BITS)
	device.Flush()
	device.Finish()
	readGPU(out, data)

	return math.Mat[T]{
		Row:  m.Row,
		Col:  m.Col,
		Data: data,
	}
}

type shaderFn struct {
	programId gles.Program
	funcAdd   computeFunc
	funcSub   computeFunc
	funcSqrt  computeFunc
}

type computeFunc struct{ progId gles.Program }

var (
	// ctx holds a EGL context. It is necessary to initialize EGL context
	// before using OpenGL ES functions.
	ctx *egl.Context

	//go:embed shaders/add.comp.glsl
	funcAdd string
	//go:embed shaders/sub.comp.glsl
	funcSub string
	//go:embed shaders/sqrt.comp.glsl
	funcSqrt string

	mathLib Function
)

func newMathShader() (f Function, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = ctx.MakeCurrent(); err != nil {
		return
	}
	defer ctx.ReleaseCurrent()

	f = Function{shaderFn{
		funcAdd:  computeFunc{try(gles.CreateComputeProgram(device, funcAdd))},
		funcSub:  computeFunc{try(gles.CreateComputeProgram(device, funcSub))},
		funcSqrt: computeFunc{try(gles.CreateComputeProgram(device, funcSqrt))},
	}}
	return
}
