// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package tests

import (
	"runtime"

	"poly.red/gpu/gl"
	"poly.red/internal/bytes"
	"poly.red/math"
)

type buffer struct {
	id     gl.Buffer
	target gl.Enum
	length int
}

func newBuffer[T DataType](target gl.Enum, length int, usageHint gl.Enum, data []T) *buffer {
	id := device.CreateBuffer()
	var zero gl.Buffer
	if id == zero {
		panic("gpu: glGenBuffers failed")
	}

	device.BindBuffer(target, id)
	device.BufferData(target, length*math.TypeSize[T](), usageHint, nil)

	buf := &buffer{
		id:     id,
		target: target,
		length: length,
	}
	device.BindBuffer(target, id)
	mbuf := device.MapBufferRange(target, 0, length*math.TypeSize[T](), gl.MAP_WRITE_BIT|gl.MAP_INVALIDATE_BUFFER_BIT)
	copy(mbuf, bytes.FromSlice(data))
	device.BindBuffer(target, id)
	if !device.UnmapBuffer(target) {
		panic("gpu: glUnmapBuffer failed")
	}

	runtime.SetFinalizer(buf, func(b *buffer) {
		var zero gl.Buffer
		if b.id == zero {
			return
		}
		device.DeleteBuffer(b.id)
		b.id = gl.Buffer{}
	})
	return buf
}

func readGPU[T DataType](b *buffer, data []T) {
	device.BindBuffer(b.target, b.id)
	buf := device.MapBufferRange(b.target, 0, b.length, gl.MAP_READ_BIT)
	copy(data, bytes.Convert[T](buf))
	device.BindBuffer(b.target, b.id)
	if !device.UnmapBuffer(b.target) {
		panic("gpu: glUnmapBuffer failed")
	}
}
