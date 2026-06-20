// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// This test closes the Go→shader loop end to end (docs/gpu-abstraction.md §6b):
// it compiles the matrix kernels from Go source, runs the emitted MSL through
// the Device API on Metal, and checks the results against the CPU math.Mat
// implementation. cgo-free.
package shader_test

import (
	"unsafe"

	"testing"

	"poly.red/gpu"
	"poly.red/gpu/shader"
	"poly.red/math"
)

const gpuEps = 1e-5

const kernelSrc = `
package kernels

type Params struct {
	WidthA uint
	HeightA uint
	WidthB uint
}

func Add(gid int, a []float32, b []float32, out []float32) { out[gid] = a[gid] + b[gid] }
func Sub(gid int, a []float32, b []float32, out []float32) { out[gid] = a[gid] - b[gid] }
func Sqrt(gid int, a []float32, out []float32) { out[gid] = sqrt(a[gid]) }

func Mul(gid int, a []float32, b []float32, out []float32, p Params) {
	row := uint(gid) / p.WidthB
	col := uint(gid) % p.WidthB
	var sum float32 = 0
	for i := uint(0); i < p.WidthA; i++ {
		sum += a[row*p.WidthA+i] * b[i*p.WidthB+col]
	}
	out[gid] = sum
}
`

func bytesOf(d []float32) []byte {
	if len(d) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

func toLayoutEntry(b shader.Binding) gpu.BindGroupLayoutEntry {
	kind := gpu.StorageBuffer
	if b.Kind == shader.UniformBuffer {
		kind = gpu.UniformBuffer
	}
	return gpu.BindGroupLayoutEntry{Binding: b.Index, Visibility: gpu.StageCompute, Kind: kind}
}

func TestGoShaderEndToEnd(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	kernels, err := shader.Compile(kernelSrc)
	if err != nil {
		t.Fatalf("compile kernels: %v", err)
	}

	// pipeline builds a compute pipeline + layout from a compiled kernel.
	build := func(name string) (*gpu.ComputePipeline, *gpu.BindGroupLayout) {
		k := kernels[name]
		mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: k.MSL})
		if err != nil {
			t.Fatalf("%s: compile MSL: %v\n%s", name, err, k.MSL)
		}
		var entries []gpu.BindGroupLayoutEntry
		for _, b := range k.SortedBindings() {
			entries = append(entries, toLayoutEntry(b))
		}
		layout := dev.NewBindGroupLayout(entries...)
		pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
			Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: name,
		})
		if err != nil {
			t.Fatalf("%s: pipeline: %v", name, err)
		}
		return pipe, layout
	}

	storage := func(d []float32) *gpu.Buffer {
		b, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: bytesOf(d)})
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	out := func(n int) *gpu.Buffer {
		b, err := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	dispatch := func(pipe *gpu.ComputePipeline, bg *gpu.BindGroup, threads, n int, ob *gpu.Buffer) []float32 {
		enc := dev.NewCommandEncoder()
		cp := enc.BeginComputePass()
		cp.SetPipeline(pipe)
		cp.SetBindGroup(0, bg)
		cp.Dispatch(threads, 1, 1)
		cp.End()
		dev.Queue().Submit(enc.Finish())
		dev.Queue().WaitIdle()
		res := make([]float32, n)
		copy(res, unsafe.Slice((*float32)(unsafe.Pointer(&ob.Bytes()[0])), n))
		return res
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	n := len(m1.Data)

	t.Run("Add", func(t *testing.T) {
		pipe, layout := build("Add")
		a, b, ob := storage(m1.Data), storage(m2.Data), out(n)
		bg := dev.NewBindGroup(layout,
			gpu.BindGroupEntry{Binding: 0, Buffer: a},
			gpu.BindGroupEntry{Binding: 1, Buffer: b},
			gpu.BindGroupEntry{Binding: 2, Buffer: ob})
		got := math.Mat[float32]{Row: 10, Col: 10, Data: dispatch(pipe, bg, n, n, ob)}
		if !got.EqEps(m1.Add(m2), gpuEps) {
			t.Fatal("Go→shader Add != CPU")
		}
	})

	t.Run("Sub", func(t *testing.T) {
		pipe, layout := build("Sub")
		a, b, ob := storage(m1.Data), storage(m2.Data), out(n)
		bg := dev.NewBindGroup(layout,
			gpu.BindGroupEntry{Binding: 0, Buffer: a},
			gpu.BindGroupEntry{Binding: 1, Buffer: b},
			gpu.BindGroupEntry{Binding: 2, Buffer: ob})
		got := math.Mat[float32]{Row: 10, Col: 10, Data: dispatch(pipe, bg, n, n, ob)}
		if !got.EqEps(m1.Sub(m2), gpuEps) {
			t.Fatal("Go→shader Sub != CPU")
		}
	})

	t.Run("Sqrt", func(t *testing.T) {
		pipe, layout := build("Sqrt")
		a, ob := storage(m1.Data), out(n)
		bg := dev.NewBindGroup(layout,
			gpu.BindGroupEntry{Binding: 0, Buffer: a},
			gpu.BindGroupEntry{Binding: 1, Buffer: ob})
		got := math.Mat[float32]{Row: 10, Col: 10, Data: dispatch(pipe, bg, n, n, ob)}
		if !got.EqEps(m1.Sqrt(), gpuEps) {
			t.Fatal("Go→shader Sqrt != CPU")
		}
	})

	t.Run("Mul", func(t *testing.T) {
		pipe, layout := build("Mul")
		a, b, ob := storage(m1.Data), storage(m2.Data), out(n)
		params := []float32{
			f32bits(uint32(m1.Col)), f32bits(uint32(m1.Row)), f32bits(uint32(m2.Col)),
		}
		pbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: 12, Usage: gpu.BufferUniform, Data: bytesOf(params)})
		if err != nil {
			t.Fatal(err)
		}
		bg := dev.NewBindGroup(layout,
			gpu.BindGroupEntry{Binding: 0, Buffer: a},
			gpu.BindGroupEntry{Binding: 1, Buffer: b},
			gpu.BindGroupEntry{Binding: 2, Buffer: ob},
			gpu.BindGroupEntry{Binding: 3, Buffer: pbuf})
		got := math.Mat[float32]{Row: 10, Col: 10, Data: dispatch(pipe, bg, m2.Col*m1.Row, n, ob)}
		if !got.EqEps(m1.Mul(m2), gpuEps) {
			t.Fatal("Go→shader Mul != CPU")
		}
	})
}

func f32bits(b uint32) float32 { return *(*float32)(unsafe.Pointer(&b)) }
