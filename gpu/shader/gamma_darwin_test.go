// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Renderer-integration slice (Phase 3 C5): the engine's sRGB gamma-correction
// pass (shader.GammaCorrection / color.FromLinear2sRGB) authored as a Go GPU
// kernel, compiled to MSL, run through the Device API on Metal, and checked
// against the engine's CPU result per channel. cgo-free.
package shader_test

import (
	"math"
	"testing"
	"unsafe"

	"poly.red/color"
	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// gammaKernel is the analytic linear->sRGB transfer (color/srgb.go:linear2sRGB)
// written in Go. The compiler now supports if/else and pow.
const gammaKernel = `
package kernels

func SRGB(gid uint, in []float32, out []float32) {
	v := in[gid]
	if v <= 0.0031308 {
		out[gid] = v * 12.92
	} else {
		out[gid] = 1.055*pow(v, 0.4166666666) - 0.055
	}
}
`

func TestGoShaderGammaMatchesEngine(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(gammaKernel)
	if err != nil {
		t.Fatalf("compile gamma kernel: %v", err)
	}
	k := ks["SRGB"]
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: k.MSL})
	if err != nil {
		t.Fatalf("MSL compile: %v\n%s", err, k.MSL)
	}
	layout := dev.NewBindGroupLayout(
		gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
		gpu.BindGroupLayoutEntry{Binding: 1, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "SRGB",
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	// Input: every normalized 8-bit level [0,1].
	const n = 256
	in := make([]float32, n)
	for i := range in {
		in[i] = float32(i) / 255
	}
	inBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: f32b(in)})
	outBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: inBuf},
		gpu.BindGroupEntry{Binding: 1, Buffer: outBuf})

	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(n, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := make([]float32, n)
	copy(got, unsafe.Slice((*float32)(unsafe.Pointer(&outBuf.Bytes()[0])), n))

	// Compare the 8-bit result to the engine's gamma per level. The engine uses
	// a LUT approximation of the same analytic curve, so allow ±1 on uint8.
	for i := 0; i < n; i++ {
		gpuU8 := int(math.Round(float64(got[i])*255 + 0.0))
		cpuU8 := int(color.FromLinear2sRGB(float32(i)/255)*255 + 0.5)
		if d := gpuU8 - cpuU8; d < -1 || d > 1 {
			t.Fatalf("level %d: GPU sRGB u8=%d, engine u8=%d (diff %d)", i, gpuU8, cpuU8, d)
		}
	}
}

func f32b(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
