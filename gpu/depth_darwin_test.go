// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Depth-buffer conformance for the Metal render pipeline (forward-rasterizer
// brick 1, specs/foundations/gpu-render-depth.md): two full-screen triangles at
// different depths are drawn in BOTH orders; with a depth attachment the nearer
// (red) triangle must win at the center regardless of draw order. The reverse
// order is the discriminating case: without depth testing the last-drawn (far,
// green) triangle would win.
package gpu_test

import (
	"testing"

	"poly.red/gpu"
)

const depthMSL = `
#include <metal_stdlib>
using namespace metal;
struct VOut { float4 pos [[position]]; float4 color; };
vertex VOut vmain(uint vid [[vertex_id]],
                  device const float* pos [[buffer(0)]],
                  device const float* col [[buffer(1)]]) {
	VOut o;
	o.pos = float4(pos[vid*3], pos[vid*3+1], pos[vid*3+2], 1.0);
	o.color = float4(col[vid*3], col[vid*3+1], col[vid*3+2], 1.0);
	return o;
}
fragment float4 fmain(VOut in [[stage_in]]) { return in.color; }
`

func TestRenderDepthOcclusion(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: depthMSL})
	if err != nil {
		t.Fatalf("compile shader: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mod, VertexEntry: "vmain",
		FragmentModule: mod, FragmentEntry: "fmain",
		ColorFormat: gpu.RGBA8Unorm,
		DepthFormat: gpu.Depth32Float,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	// Metal NDC depth is [0,1]; with compare "less" a smaller z is nearer.
	// A full-screen triangle (the big-triangle trick) at z=0.2 (near, red) and
	// z=0.8 (far, green); both cover the center pixel.
	near := []float32{-1, -1, 0.2, 3, -1, 0.2, -1, 3, 0.2}
	far := []float32{-1, -1, 0.8, 3, -1, 0.8, -1, 3, 0.8}
	red := []float32{1, 0, 0, 1, 0, 0, 1, 0, 0}
	green := []float32{0, 1, 0, 0, 1, 0, 0, 1, 0}

	buf := func(d []float32) *gpu.Buffer {
		b, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(d) * 4, Usage: gpu.BufferStorage, Data: bytesFromFloats(d)})
		if err != nil {
			t.Fatalf("buffer: %v", err)
		}
		return b
	}
	nearPos, farPos, redCol, greenCol := buf(near), buf(far), buf(red), buf(green)

	// render draws the two triangles in the given order into a fresh color+depth
	// target and returns the center pixel.
	render := func(firstFar bool) (r, g, b uint8) {
		color, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})
		if err != nil {
			t.Fatalf("color texture: %v", err)
		}
		depth, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.Depth32Float, Width: W, Height: H, RenderTarget: true})
		if err != nil {
			t.Fatalf("depth texture: %v", err)
		}
		enc := dev.NewCommandEncoder()
		rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
			ColorTexture: color, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 1, 1}, // blue
			DepthTexture: depth, ClearDepth: 1,
		})
		rp.SetPipeline(pipe)
		draw := func(pos, col *gpu.Buffer) {
			rp.SetVertexBuffer(0, pos)
			rp.SetVertexBuffer(1, col)
			rp.Draw(gpu.TriangleList, 0, 3)
		}
		if firstFar {
			draw(farPos, greenCol)
			draw(nearPos, redCol)
		} else {
			draw(nearPos, redCol)
			draw(farPos, greenCol)
		}
		rp.End()
		dev.Queue().Submit(enc.Finish())
		dev.Queue().WaitIdle()
		px := color.ReadPixels()
		i := (H/2*W + W/2) * 4
		return px[i], px[i+1], px[i+2]
	}

	for _, firstFar := range []bool{true, false} {
		r, g, b := render(firstFar)
		if r < 200 || g > 60 || b > 60 {
			t.Errorf("firstFar=%v: center = (%d,%d,%d), want red (near triangle wins via depth test)", firstFar, r, g, b)
		}
	}
}
