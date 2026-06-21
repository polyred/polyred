// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Conformance test for the cgo-free OpenGL ES compute backend, driven entirely
// through the public Device API (the same surface the Metal backend serves).
// Kernels are authored in Go, compiled to GLSL by shader.CompileGLSL, and their
// results checked against the CPU. Runs in CI on Mesa llvmpipe (software), gated
// on EGL_PLATFORM=surfaceless; a no-op on machines without a GL device.
package gpu_test

import (
	"math"
	"os"
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

func glBytesOf(d []float32) []byte {
	if len(d) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

func glFloatsOf(b []byte, n int) []float32 {
	out := make([]float32, n)
	copy(out, unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), n))
	return out
}

// runGLKernel compiles src (Go -> GLSL), runs entry over n threads with the
// given storage buffers (in declaration order), and returns the contents of the
// output buffer named by outIdx.
func runGLKernel(t *testing.T, dev *gpu.Device, src, entry string, n int, bufs [][]float32, outIdx int) []float32 {
	t.Helper()
	ks, err := shader.CompileGLSL(src)
	if err != nil {
		t.Fatalf("CompileGLSL %s: %v", entry, err)
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: ks[entry].GLSL})
	if err != nil {
		t.Fatalf("NewShaderModule %s: %v", entry, err)
	}
	var entries []gpu.BindGroupLayoutEntry
	for i := range bufs {
		entries = append(entries, gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer})
	}
	layout := dev.NewBindGroupLayout(entries...)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: entry,
	})
	if err != nil {
		t.Fatalf("NewComputePipeline %s: %v", entry, err)
	}

	gpuBufs := make([]*gpu.Buffer, len(bufs))
	var bgEntries []gpu.BindGroupEntry
	for i, b := range bufs {
		gb, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(b), Usage: gpu.BufferStorage})
		if err != nil {
			t.Fatalf("NewBuffer %d: %v", i, err)
		}
		gpuBufs[i] = gb
		bgEntries = append(bgEntries, gpu.BindGroupEntry{Binding: i, Buffer: gb})
	}
	bg := dev.NewBindGroup(layout, bgEntries...)

	enc := dev.NewCommandEncoder()
	pass := enc.BeginComputePass()
	pass.SetPipeline(pipe)
	pass.SetBindGroup(0, bg)
	pass.Dispatch(n, 1, 1)
	pass.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	return glFloatsOf(gpuBufs[outIdx].Bytes(), n)
}

func TestGLBackendCompute(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL backend conformance test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()
	if dev.Driver() != gpu.DriverGL {
		t.Fatalf("expected DriverGL, got %v", dev.Driver())
	}

	const n = 1024
	a := make([]float32, n)
	b := make([]float32, n)
	for i := range a {
		a[i] = float32(i) * 0.5
		b[i] = float32(n - i)
	}

	const src = `package kernels
func Add(gid uint, a []float32, b []float32, out []float32)  { out[gid] = a[gid] + b[gid] }
func Sub(gid uint, a []float32, b []float32, out []float32)  { out[gid] = a[gid] - b[gid] }
func Sqrt(gid uint, a []float32, out []float32)              { out[gid] = sqrt(a[gid]) }`

	t.Run("Add", func(t *testing.T) {
		got := runGLKernel(t, dev, src, "Add", n, [][]float32{a, b, make([]float32, n)}, 2)
		for i := range got {
			if got[i] != a[i]+b[i] {
				t.Fatalf("Add[%d] = %v, want %v", i, got[i], a[i]+b[i])
			}
		}
	})
	t.Run("Sub", func(t *testing.T) {
		got := runGLKernel(t, dev, src, "Sub", n, [][]float32{a, b, make([]float32, n)}, 2)
		for i := range got {
			if got[i] != a[i]-b[i] {
				t.Fatalf("Sub[%d] = %v, want %v", i, got[i], a[i]-b[i])
			}
		}
	})
	t.Run("Sqrt", func(t *testing.T) {
		got := runGLKernel(t, dev, src, "Sqrt", n, [][]float32{a, make([]float32, n)}, 1)
		for i := range got {
			want := float32(math.Sqrt(float64(a[i])))
			d := got[i] - want
			if d < 0 {
				d = -d
			}
			if d > 1e-5 {
				t.Fatalf("Sqrt[%d] = %v, want ~%v", i, got[i], want)
			}
		}
	})
}
