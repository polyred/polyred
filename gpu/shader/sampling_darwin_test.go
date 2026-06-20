// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Texture sampling authored in Go: a compute kernel takes a Texture2D + Sampler,
// samples the four texel centres of a 2x2 texture through the Device API, and
// returns the colours. Proves Go→shader texture sampling end to end, cgo-free —
// the foundation for textured shading and shadow-map sampling.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

const samplingKernel = `
package kernels

type Vec2 struct{ X, Y float32 }
type Vec4 struct{ X, Y, Z, W float32 }

func SampleTex(gid uint, tex Texture2D, samp Sampler, out []float32) {
	u := float32(gid%2)*0.5 + 0.25
	v := float32(gid/2)*0.5 + 0.25
	c := tex.Sample(samp, Vec2{u, v})
	out[gid*4] = c.X
	out[gid*4+1] = c.Y
	out[gid*4+2] = c.Z
	out[gid*4+3] = c.W
}
`

func TestGoShaderTextureSampling(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(samplingKernel)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	k := ks["SampleTex"]
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: k.MSL})
	if err != nil {
		t.Fatalf("MSL: %v\n%s", err, k.MSL)
	}
	// out is the lone storage buffer at binding 0 (tex/sampler use other spaces).
	layout := dev.NewBindGroupLayout(gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer})
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "SampleTex"})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	tex, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: 2, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	tex.Write([]byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 255, 255,
	})
	samp := dev.NewSampler(gpu.SamplerDescriptor{MinFilter: gpu.FilterNearest, MagFilter: gpu.FilterNearest})

	out, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 4 * 4 * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	bg := dev.NewBindGroup(layout, gpu.BindGroupEntry{Binding: 0, Buffer: out})

	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetTexture(0, tex)
	cp.SetSampler(0, samp)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(4, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := make([]float32, 16)
	copy(got, unsafe.Slice((*float32)(unsafe.Pointer(&out.Bytes()[0])), 16))
	want := [][3]float32{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {1, 1, 1}}
	for i, w := range want {
		r, g, b := got[i*4], got[i*4+1], got[i*4+2]
		if abs(r-w[0]) > 0.02 || abs(g-w[1]) > 0.02 || abs(b-w[2]) > 0.02 {
			t.Fatalf("texel %d: got (%.2f,%.2f,%.2f) want %v\nMSL:\n%s", i, r, g, b, w, k.MSL)
		}
	}
}

func abs(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
