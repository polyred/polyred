// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package gpu

import (
	_ "embed"
	"errors"
	"runtime"

	"fmt"

	"poly.red/internal/bytes"
	"poly.red/internal/driver/egl"
	"poly.red/internal/driver/gles"
	"poly.red/math"
)

var device *gles.Functions

func init() {
	ctx = try(egl.NewContext(egl.NewDisplay()))
	device = try(gles.NewFunctions())
	addFn = try(newAddShader())
}

// add is a GPU version of math.Mat[float32].Add method.
func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ctx.MakeCurrent(); err != nil {
		panic("gpu: cannot establish egl context")
	}
	defer ctx.MakeCurrent()

	data := make([]T, len(m1.Data))
	out := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m1.Data), math.TypeSize[T](), gles.STREAM_COPY, data)
	bufA := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m1.Data), math.TypeSize[T](), gles.STREAM_COPY, m1.Data)
	bufB := newBuffer(gles.SHADER_STORAGE_BUFFER, len(m1.Data), math.TypeSize[T](), gles.STREAM_COPY, m2.Data)

	device.BindBufferBase(out.target, 0, out.id)
	device.BindBufferBase(bufA.target, 1, bufA.id)
	device.BindBufferBase(bufB.target, 2, bufB.id)

	device.UseProgram(addFn.programId)
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

type buffer struct {
	id     gles.Buffer
	target gles.Enum
	length int
}

func newBuffer[T DataType](target gles.Enum, length int, elemSize int, usageHint gles.Enum, data []T) *buffer {
	id := device.CreateBuffer()
	var zero gles.Buffer
	if id == zero {
		panic("gpu: glGenBuffers failed")
	}

	device.BindBuffer(target, id)
	device.BufferData(target, length*elemSize, usageHint, nil)

	buf := &buffer{
		id:     id,
		target: target,
		length: length,
	}
	device.BindBuffer(target, id)
	mbuf := device.MapBufferRange(target, 0, length*elemSize, gles.MAP_WRITE_BIT|gles.MAP_INVALIDATE_BUFFER_BIT)
	copy(mbuf, bytes.FromSlice(data))
	device.BindBuffer(target, id)
	if !device.UnmapBuffer(target) {
		panic("gpu: glUnmapBuffer failed")
	}

	runtime.SetFinalizer(buf, func(b *buffer) {
		var zero gles.Buffer
		if b.id == zero {
			return
		}
		device.DeleteBuffer(b.id)
		b.id = gles.Buffer{}
	})
	return buf
}

func readGPU[T DataType](b *buffer, data []T) {
	device.BindBuffer(b.target, b.id)
	buf := device.MapBufferRange(b.target, 0, b.length, gles.MAP_READ_BIT)
	copy(data, bytes.Convert[T](buf))
	device.BindBuffer(b.target, b.id)
	if !device.UnmapBuffer(b.target) {
		panic("gpu glUnmapBuffer failed")
	}
}

type shaderFn struct {
	programId gles.Program
}

var (
	// ctx holds a EGL context. It is necessary to initialize EGL context
	// before using OpenGL ES functions.
	ctx *egl.Context

	//go:embed add.comp.glsl
	addGl string
	addFn Function
)

func newAddShader() (sf Function, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err = ctx.MakeCurrent(); err != nil {
		return
	}
	defer ctx.ReleaseCurrent()

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
