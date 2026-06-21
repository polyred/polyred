// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Multiple-render-target conformance for the Metal render pipeline
// (forward-rasterizer brick 2, specs/foundations/gpu-render-mrt.md): one draw
// whose fragment shader writes a distinct color to each of three attachments;
// each target is read back and must hold its own color (not aliased, not only
// attachment 0). This is the G-buffer enabler the deferred path needs.
package gpu_test

import (
	"testing"

	"poly.red/gpu"
)

const mrtMSL = `
#include <metal_stdlib>
using namespace metal;
struct FOut {
	float4 c0 [[color(0)]];
	float4 c1 [[color(1)]];
	float4 c2 [[color(2)]];
};
vertex float4 vmain(uint vid [[vertex_id]], device const float* pos [[buffer(0)]]) {
	return float4(pos[vid*2], pos[vid*2+1], 0.0, 1.0);
}
fragment FOut fmain() {
	FOut o;
	o.c0 = float4(1.0, 0.0, 0.0, 1.0); // red
	o.c1 = float4(0.0, 1.0, 0.0, 1.0); // green
	o.c2 = float4(0.0, 0.0, 1.0, 1.0); // blue
	return o;
}
`

func TestRenderMRT(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: mrtMSL})
	if err != nil {
		t.Fatalf("compile shader: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mod, VertexEntry: "vmain",
		FragmentModule: mod, FragmentEntry: "fmain",
		ColorFormat:       gpu.RGBA8Unorm,
		ExtraColorFormats: []gpu.TextureFormat{gpu.RGBA8Unorm, gpu.RGBA8Unorm},
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	tex := func() *gpu.Texture {
		x, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})
		if err != nil {
			t.Fatalf("texture: %v", err)
		}
		return x
	}
	t0, t1, t2 := tex(), tex(), tex()

	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(verts) * 4, Usage: gpu.BufferStorage, Data: bytesFromFloats(verts)})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: t0, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
		ExtraColorTargets: []gpu.ColorTarget{
			{Texture: t1, ClearColor: [4]float64{0, 0, 0, 1}},
			{Texture: t2, ClearColor: [4]float64{0, 0, 0, 1}},
		},
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	center := func(x *gpu.Texture) (r, g, b uint8) {
		px := x.ReadPixels()
		i := (H/2*W + W/2) * 4
		return px[i], px[i+1], px[i+2]
	}
	want := []struct {
		name    string
		tex     *gpu.Texture
		r, g, b uint8
	}{
		{"target0", t0, 255, 0, 0}, // red
		{"target1", t1, 0, 255, 0}, // green
		{"target2", t2, 0, 0, 255}, // blue
	}
	for _, w := range want {
		r, g, b := center(w.tex)
		if absDiff(r, w.r) > 8 || absDiff(g, w.g) > 8 || absDiff(b, w.b) > 8 {
			t.Errorf("%s center = (%d,%d,%d), want (%d,%d,%d)", w.name, r, g, b, w.r, w.g, w.b)
		}
	}
}

func absDiff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
