// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package gpu

import (
	"errors"
	"unsafe"

	"poly.red/gpu/mtl"
)

// openBackend selects the Metal backend on darwin.
func openBackend(d Driver) (backend, Driver, error) {
	switch d {
	case DriverAuto, DriverMetal:
		dev, err := mtl.CreateSystemDefaultDevice()
		if err != nil || !dev.Available() {
			return nil, DriverAuto, ErrUnsupported
		}
		return &metalBackend{dev: dev, queue: dev.MakeCommandQueue()}, DriverMetal, nil
	default:
		return nil, DriverAuto, ErrUnsupported
	}
}

type metalBackend struct {
	dev     mtl.Device
	queue   mtl.CommandQueue
	last    mtl.CommandBuffer // last committed buffer, for waitIdle
	hasLast bool
}

func (m *metalBackend) newBuffer(size int, usage BufferUsage, data []byte) (backendBuffer, error) {
	var buf mtl.Buffer
	if len(data) > 0 {
		buf = m.dev.MakeBuffer(unsafe.Pointer(&data[0]), uintptr(size), mtl.ResourceStorageModeShared)
	} else {
		buf = m.dev.MakeBuffer(nil, uintptr(size), mtl.ResourceStorageModeShared)
	}
	return &metalBuffer{buf: buf, size: size}, nil
}

func (m *metalBackend) newShaderModule(src ShaderSource) (backendShaderModule, error) {
	if src.MSL == "" {
		return nil, errors.New("gpu: metal backend requires ShaderSource.MSL")
	}
	lib, err := m.dev.MakeLibrary(src.MSL, mtl.CompileOptions{})
	if err != nil {
		return nil, err
	}
	return &metalModule{lib: lib}, nil
}

func (m *metalBackend) newComputePipeline(mod backendShaderModule, entry string) (backendComputePipeline, error) {
	mm, ok := mod.(*metalModule)
	if !ok {
		return nil, errors.New("gpu: shader module is not a metal module")
	}
	fn, err := mm.lib.MakeFunction(entry)
	if err != nil {
		return nil, err
	}
	cps, err := m.dev.MakeComputePipelineState(fn)
	if err != nil {
		return nil, err
	}
	return &metalPipeline{cps: cps, max: cps.MaxTotalThreadsPerThreadgroup()}, nil
}

func (m *metalBackend) newCommandBuffer() backendCommandBuffer {
	return &metalCmd{m: m, cb: m.queue.MakeCommandBuffer()}
}

func (m *metalBackend) waitIdle() {
	if m.hasLast {
		m.last.WaitUntilCompleted()
	}
}

func (m *metalBackend) close() error {
	m.queue.Release()
	return nil
}

type metalBuffer struct {
	buf  mtl.Buffer
	size int
}

func (b *metalBuffer) bytes() []byte {
	return unsafe.Slice((*byte)(b.buf.Content()), b.size)
}
func (b *metalBuffer) release() { b.buf.Release() }

type metalModule struct{ lib mtl.Library }

func (*metalModule) isShaderModule() {}

type metalPipeline struct {
	cps mtl.ComputePipelineState
	max int
}

func (p *metalPipeline) maxThreads() int { return p.max }

type metalCmd struct {
	m   *metalBackend
	cb  mtl.CommandBuffer
	enc mtl.ComputeCommandEncoder
	cur *metalPipeline
}

func (c *metalCmd) beginCompute() { c.enc = c.cb.MakeComputeCommandEncoder() }

func (c *metalCmd) setComputePipeline(p backendComputePipeline) {
	mp := p.(*metalPipeline)
	c.cur = mp
	c.enc.SetComputePipelineState(mp.cps)
}

func (c *metalCmd) setBuffer(b backendBuffer, offset, index int) {
	c.enc.SetBuffer(b.(*metalBuffer).buf, offset, index)
}

func (c *metalCmd) dispatch(x, y, z int) {
	total := x
	if y > 1 {
		total *= y
	}
	if z > 1 {
		total *= z
	}
	tg := c.cur.max
	if tg > total {
		tg = total
	}
	if tg < 1 {
		tg = 1
	}
	c.enc.DispatchThreads(
		mtl.Size{Width: x, Height: y, Depth: z},
		mtl.Size{Width: tg, Height: 1, Depth: 1},
	)
}

func (c *metalCmd) endCompute() { c.enc.EndEncoding() }

func (c *metalCmd) commit() {
	c.cb.Commit()
	c.m.last = c.cb
	c.m.hasLast = true
}
