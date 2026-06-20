// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// This test is the Phase 1 compute vertical slice (docs/gpu-abstraction.md
// §5a): it reimplements the matrix Add/Sub/Sqrt/Mul GPU kernels through the
// backend-agnostic Device API (instead of talking to mtl directly) and checks
// the results against the CPU math.Mat implementation. It runs cgo-free.
package gpu_test

import (
	_ "embed"
	"unsafe"

	"testing"

	"poly.red/gpu"
	"poly.red/math"
)

//go:embed tests/shaders/math.metal
var mathMSL string

// gpuEps is a float32-appropriate tolerance for comparing GPU and CPU results
// (see math.Mat.EqEps and gpu/tests/math_test.go).
const gpuEps = 1e-5

func bytesOf(data []float32) []byte {
	if len(data) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&data[0])), len(data)*4)
}

func floatsOf(b []byte, n int) []float32 {
	out := make([]float32, n)
	copy(out, unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), n))
	return out
}

// harness opens a device and compiles the math library once per test.
type harness struct {
	dev  *gpu.Device
	mod  *gpu.ShaderModule
	lay2 *gpu.BindGroupLayout // a, out
	lay3 *gpu.BindGroupLayout // a, b, out
	lay4 *gpu.BindGroupLayout // a, b, out, params
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device available: %v", err)
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: mathMSL})
	if err != nil {
		t.Fatalf("compile math.metal: %v", err)
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	return &harness{
		dev:  dev,
		mod:  mod,
		lay2: dev.NewBindGroupLayout(sb(0), sb(1)),
		lay3: dev.NewBindGroupLayout(sb(0), sb(1), sb(2)),
		lay4: dev.NewBindGroupLayout(sb(0), sb(1), sb(2),
			gpu.BindGroupLayoutEntry{Binding: 3, Visibility: gpu.StageCompute, Kind: gpu.UniformBuffer}),
	}
}

func (h *harness) buf(t *testing.T, data []float32) *gpu.Buffer {
	t.Helper()
	b, err := h.dev.NewBuffer(gpu.BufferDescriptor{Size: len(data) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: bytesOf(data)})
	if err != nil {
		t.Fatalf("new buffer: %v", err)
	}
	return b
}

func (h *harness) outBuf(t *testing.T, n int) *gpu.Buffer {
	t.Helper()
	b, err := h.dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	if err != nil {
		t.Fatalf("new out buffer: %v", err)
	}
	return b
}

// run dispatches entry over threads using the given bind group and returns the
// out buffer contents as a float32 slice of length n.
func (h *harness) run(t *testing.T, entry string, layout *gpu.BindGroupLayout, bg *gpu.BindGroup, threads, n int, out *gpu.Buffer) []float32 {
	t.Helper()
	pl := h.dev.NewPipelineLayout(layout)
	pipe, err := h.dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: pl, Module: h.mod, Entry: entry})
	if err != nil {
		t.Fatalf("pipeline %s: %v", entry, err)
	}
	enc := h.dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(threads, 1, 1)
	cp.End()
	h.dev.Queue().Submit(enc.Finish())
	h.dev.Queue().WaitIdle()
	return floatsOf(out.Bytes(), n)
}

func TestDeviceCompute_Add(t *testing.T) {
	h := newHarness(t)
	defer h.dev.Close()
	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	n := len(m1.Data)
	a, b, out := h.buf(t, m1.Data), h.buf(t, m2.Data), h.outBuf(t, n)
	bg := h.dev.NewBindGroup(h.lay3,
		gpu.BindGroupEntry{Binding: 0, Buffer: a},
		gpu.BindGroupEntry{Binding: 1, Buffer: b},
		gpu.BindGroupEntry{Binding: 2, Buffer: out})
	got := math.Mat[float32]{Row: 10, Col: 10, Data: h.run(t, "add0", h.lay3, bg, n, n, out)}
	if !got.EqEps(m1.Add(m2), gpuEps) {
		t.Fatalf("Add through Device API != CPU\nGPU=%v\nCPU=%v", got, m1.Add(m2))
	}
}

func TestDeviceCompute_Sub(t *testing.T) {
	h := newHarness(t)
	defer h.dev.Close()
	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	n := len(m1.Data)
	a, b, out := h.buf(t, m1.Data), h.buf(t, m2.Data), h.outBuf(t, n)
	bg := h.dev.NewBindGroup(h.lay3,
		gpu.BindGroupEntry{Binding: 0, Buffer: a},
		gpu.BindGroupEntry{Binding: 1, Buffer: b},
		gpu.BindGroupEntry{Binding: 2, Buffer: out})
	got := math.Mat[float32]{Row: 10, Col: 10, Data: h.run(t, "sub0", h.lay3, bg, n, n, out)}
	if !got.EqEps(m1.Sub(m2), gpuEps) {
		t.Fatalf("Sub through Device API != CPU")
	}
}

func TestDeviceCompute_Sqrt(t *testing.T) {
	h := newHarness(t)
	defer h.dev.Close()
	m1 := math.NewRandMat[float32](10, 10)
	n := len(m1.Data)
	a, out := h.buf(t, m1.Data), h.outBuf(t, n)
	bg := h.dev.NewBindGroup(h.lay2,
		gpu.BindGroupEntry{Binding: 0, Buffer: a},
		gpu.BindGroupEntry{Binding: 1, Buffer: out})
	got := math.Mat[float32]{Row: 10, Col: 10, Data: h.run(t, "sqrt0", h.lay2, bg, n, n, out)}
	if !got.EqEps(m1.Sqrt(), gpuEps) {
		t.Fatalf("Sqrt through Device API != CPU")
	}
}

func TestDeviceCompute_Mul(t *testing.T) {
	h := newHarness(t)
	defer h.dev.Close()
	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	n := len(m1.Data)
	a, b, out := h.buf(t, m1.Data), h.buf(t, m2.Data), h.outBuf(t, n)
	// params: widthA, heightA, widthB (matches struct Params in math.metal)
	params := []float32{
		float32frombits(uint32(m1.Col)),
		float32frombits(uint32(m1.Row)),
		float32frombits(uint32(m2.Col)),
	}
	pbuf, err := h.dev.NewBuffer(gpu.BufferDescriptor{Size: 12, Usage: gpu.BufferUniform, Data: bytesOf(params)})
	if err != nil {
		t.Fatalf("params buffer: %v", err)
	}
	bg := h.dev.NewBindGroup(h.lay4,
		gpu.BindGroupEntry{Binding: 0, Buffer: a},
		gpu.BindGroupEntry{Binding: 1, Buffer: b},
		gpu.BindGroupEntry{Binding: 2, Buffer: out},
		gpu.BindGroupEntry{Binding: 3, Buffer: pbuf})
	got := math.Mat[float32]{Row: 10, Col: 10, Data: h.run(t, "mul0", h.lay4, bg, m2.Col*m1.Row, n, out)}
	if !got.EqEps(m1.Mul(m2), gpuEps) {
		t.Fatalf("Mul through Device API != CPU")
	}
}

// float32frombits reinterprets a uint32 bit pattern as float32, so that the
// uint params in math.metal's Params struct are uploaded with the right bytes
// via a []float32-backed buffer.
func float32frombits(b uint32) float32 {
	return *(*float32)(unsafe.Pointer(&b))
}
