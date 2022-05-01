// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package gpu

import (
	"runtime"

	"poly.red/internal/bytes"
	"poly.red/internal/driver/gles"
	"poly.red/math"
)

type buffer struct {
	id     gles.Buffer
	target gles.Enum
	length int
}

func newBuffer[T DataType](target gles.Enum, length int, usageHint gles.Enum, data []T) *buffer {
	id := device.CreateBuffer()
	var zero gles.Buffer
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
	mbuf := device.MapBufferRange(target, 0, length*math.TypeSize[T](), gles.MAP_WRITE_BIT|gles.MAP_INVALIDATE_BUFFER_BIT)
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
		panic("gpu: glUnmapBuffer failed")
	}
}
