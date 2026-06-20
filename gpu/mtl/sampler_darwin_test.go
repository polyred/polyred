// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package mtl_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu/mtl"
)

const samplerKernel = `
#include <metal_stdlib>
using namespace metal;
kernel void samp(texture2d<float> tex [[texture(0)]],
                 sampler s [[sampler(0)]],
                 device float4* out [[buffer(0)]],
                 uint gid [[thread_position_in_grid]]) {
	// Sample the four texel centres of a 2x2 texture.
	float2 uv = float2((gid % 2) * 0.5 + 0.25, (gid / 2) * 0.5 + 0.25);
	out[gid] = tex.sample(s, uv);
}
`

// TestComputeTextureSampling drives texture sampling inside a compute kernel
// (texture binding + sampler state) end to end, cgo-free. This is the
// foundation for sampling shadow maps / textures in shaders.
func TestComputeTextureSampling(t *testing.T) {
	dev, err := mtl.CreateSystemDefaultDevice()
	if err != nil || !dev.Available() {
		t.Skip("no Metal device")
	}

	// 2x2 RGBA8 texture: red, green, blue, white.
	tex := dev.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatRGBA8UNorm, Width: 2, Height: 2,
		StorageMode: mtl.StorageModeShared, Usage: mtl.TextureUsageShaderRead,
	})
	pix := []byte{
		255, 0, 0, 255, 0, 255, 0, 255,
		0, 0, 255, 255, 255, 255, 255, 255,
	}
	tex.ReplaceRegion(mtl.RegionMake2D(0, 0, 2, 2), 0, pix, 2*4)

	samp := dev.MakeSamplerState(mtl.SamplerDescriptor{
		MinFilter: mtl.SamplerFilterNearest, MagFilter: mtl.SamplerFilterNearest,
		SAddressMode: mtl.SamplerAddressClampToEdge, TAddressMode: mtl.SamplerAddressClampToEdge,
	})

	lib, err := dev.MakeLibrary(samplerKernel, mtl.CompileOptions{})
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	fn, err := lib.MakeFunction("samp")
	if err != nil {
		t.Fatal(err)
	}
	pso, err := dev.MakeComputePipelineState(fn)
	if err != nil {
		t.Fatal(err)
	}

	out := dev.MakeBuffer(nil, uintptr(4*4*4), mtl.ResourceStorageModeShared) // 4 float4

	cq := dev.MakeCommandQueue()
	cb := cq.MakeCommandBuffer()
	enc := cb.MakeComputeCommandEncoder()
	enc.SetComputePipelineState(pso)
	enc.SetTexture(tex, 0)
	enc.SetSamplerState(samp, 0)
	enc.SetBuffer(out, 0, 0)
	enc.DispatchThreads(mtl.Size{Width: 4, Height: 1, Depth: 1}, mtl.Size{Width: 4, Height: 1, Depth: 1})
	enc.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	got := unsafe.Slice((*float32)(out.Content()), 16)
	// Expected per-texel RGB (normalized): red, green, blue, white.
	want := [][3]float32{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}, {1, 1, 1}}
	for i, w := range want {
		r, g, b := got[i*4], got[i*4+1], got[i*4+2]
		if abs32(r-w[0]) > 0.02 || abs32(g-w[1]) > 0.02 || abs32(b-w[2]) > 0.02 {
			t.Fatalf("texel %d: got (%.2f,%.2f,%.2f), want %v", i, r, g, b, w)
		}
	}
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
