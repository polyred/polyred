// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Phase 3 render slice (docs/gpu-abstraction.md §5b, specs/foundations/
// gpu-phase3-render.md): render a triangle headless to an offscreen RGBA
// texture through the Device API on Metal, read pixels back, and assert the
// result. cgo-free. Vertex/fragment kernels are hand-written MSL here (the
// Go→shader vertex/fragment profile is a later step).
package gpu_test

import (
	"image"
	"image/color"
	"testing"
	"unsafe"

	"poly.red/gpu"
)

const triangleMSL = `
#include <metal_stdlib>
using namespace metal;
struct VOut { float4 pos [[position]]; };
vertex VOut vmain(uint vid [[vertex_id]], device const float2* verts [[buffer(0)]]) {
	VOut o;
	o.pos = float4(verts[vid], 0.0, 1.0);
	return o;
}
fragment float4 fmain() { return float4(1.0, 0.0, 0.0, 1.0); } // red
`

func TestRenderTriangleHeadless(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: triangleMSL})
	if err != nil {
		t.Fatalf("compile shader: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mod, VertexEntry: "vmain",
		FragmentModule: mod, FragmentEntry: "fmain",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	target, err := dev.NewTexture(gpu.TextureDescriptor{
		Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true,
	})
	if err != nil {
		t.Fatalf("texture: %v", err)
	}

	// A triangle that covers the whole clip space, so the center is red.
	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{
		Size: len(verts) * 4, Usage: gpu.BufferStorage, Data: bytesFromFloats(verts),
	})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: target,
		Load:         gpu.LoadClear,
		ClearColor:   [4]float64{0, 1, 0, 1}, // green
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	pixels := target.ReadPixels()
	if len(pixels) != W*H*4 {
		t.Fatalf("readback size = %d, want %d", len(pixels), W*H*4)
	}

	// Center pixel must be the red triangle.
	at := func(x, y int) color.RGBA {
		i := (y*W + x) * 4
		return color.RGBA{R: pixels[i], G: pixels[i+1], B: pixels[i+2], A: pixels[i+3]}
	}
	center := at(W/2, H/2)
	if center.R < 200 || center.G > 60 || center.B > 60 {
		t.Fatalf("center pixel = %+v, want red", center)
	}

	// Sanity: produce an image so the result is inspectable if needed.
	img := image.NewRGBA(image.Rect(0, 0, W, H))
	copy(img.Pix, pixels)
	_ = img
}

func bytesFromFloats(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
