// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Multiple-render-target conformance for the GL backend (forward-rasterizer brick
// 3a, the GL CI counterpart of the Metal-only gpu-render-mrt.md): a fragment
// shader writes two outputs into two color attachments of one pass; each
// attachment must receive its own output (red to 0, green to 1). This is the
// G-buffer plumbing the GPU forward rasterizer writes normals/worldpos/basecol
// into. Runs in CI on Mesa llvmpipe (software, surfaceless).
package gpu_test

import (
	"os"
	"testing"

	"poly.red/gpu"
)

const mrtGLVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _v { float verts[]; };
void main() {
	gl_Position = vec4(verts[gl_VertexID*2], verts[gl_VertexID*2+1], 0.0, 1.0);
}`

const mrtGLFrag = `#version 310 es
precision highp float;
layout(location = 0) out vec4 out0;
layout(location = 1) out vec4 out1;
void main() {
	out0 = vec4(1.0, 0.0, 0.0, 1.0); // red  -> attachment 0
	out1 = vec4(0.0, 1.0, 0.0, 1.0); // green -> attachment 1
}`

func TestGLRenderMRT(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL MRT test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: mrtGLVert})
	if err != nil {
		t.Fatalf("vertex module: %v", err)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: mrtGLFrag})
	if err != nil {
		t.Fatalf("fragment module: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat:       gpu.RGBA8Unorm,
		ExtraColorFormats: []gpu.TextureFormat{gpu.RGBA8Unorm},
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	mkTex := func() *gpu.Texture {
		tex, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})
		if err != nil {
			t.Fatalf("texture: %v", err)
		}
		return tex
	}
	tex0, tex1 := mkTex(), mkTex()

	verts := []float32{-1, -1, 3, -1, -1, 3} // full-screen triangle
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(verts), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: tex0, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
		ExtraColorTargets: []gpu.ColorTarget{{Texture: tex1, ClearColor: [4]float64{0, 0, 0, 1}}},
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	c := ((H/2)*W + W/2) * 4
	p0 := tex0.ReadPixels()
	p1 := tex1.ReadPixels()
	if p0[c] < 200 || p0[c+1] > 60 {
		t.Fatalf("attachment 0 center = (%d,%d,%d), want red", p0[c], p0[c+1], p0[c+2])
	}
	if p1[c+1] < 200 || p1[c] > 60 {
		t.Fatalf("attachment 1 center = (%d,%d,%d), want green", p1[c], p1[c+1], p1[c+2])
	}
}
