// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Float render-target conformance for the GL backend (forward-rasterizer brick:
// the G-buffer stores world positions and normals at full precision, which 8-bit
// RGBA cannot hold). A fragment shader writes known out-of-[0,1] float values into
// an RGBA32F target; reading them back must reproduce them exactly. Runs on Mesa
// llvmpipe (surfaceless).
package gpu_test

import (
	"encoding/binary"
	stdmath "math"
	"os"
	"testing"

	"poly.red/gpu"
)

const floatTgtVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _v { float verts[]; };
void main() {
	gl_Position = vec4(verts[gl_VertexID*2], verts[gl_VertexID*2+1], 0.0, 1.0);
}`

const floatTgtFrag = `#version 310 es
precision highp float;
out vec4 fragColor;
void main() { fragColor = vec4(1.5, -2.5, 100.0, 0.25); }`

func TestGLFloatRenderTarget(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL float-target test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: floatTgtVert})
	if err != nil {
		t.Fatalf("vertex module: %v", err)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: floatTgtFrag})
	if err != nil {
		t.Fatalf("fragment module: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat: gpu.RGBA32Float,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 8, 8
	tex, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA32Float, Width: W, Height: H, RenderTarget: true})
	if err != nil {
		t.Fatalf("texture: %v", err)
	}
	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(verts), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: tex, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 0},
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	// Readback is 16 bytes/pixel (4 float32). Decode the center pixel.
	raw := tex.ReadPixels()
	if len(raw) != W*H*16 {
		t.Fatalf("ReadPixels returned %d bytes, want %d (RGBA32F)", len(raw), W*H*16)
	}
	off := ((H/2)*W + W/2) * 16
	got := [4]float32{
		f32(raw[off:]), f32(raw[off+4:]), f32(raw[off+8:]), f32(raw[off+12:]),
	}
	want := [4]float32{1.5, -2.5, 100.0, 0.25}
	for i := range want {
		if d := got[i] - want[i]; d > 1e-4 || d < -1e-4 {
			t.Fatalf("float target center = %v, want %v (channel %d off)", got, want, i)
		}
	}
}

func f32(b []byte) float32 {
	return stdmath.Float32frombits(binary.LittleEndian.Uint32(b))
}
