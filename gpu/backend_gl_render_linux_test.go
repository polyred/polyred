// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Conformance test for the GL backend's render-to-texture path: it renders a
// full-screen triangle into an offscreen texture and reads the pixels back, all
// cgo-free through the public Device API. Vertices come from a storage buffer
// indexed by gl_VertexID (mirroring the Metal vertex model). Runs in CI on Mesa
// llvmpipe (software, surfaceless). This is the headless render path that
// windowed present (gpu-windowed-present.md) builds on.
package gpu_test

import (
	"os"
	"testing"

	"poly.red/gpu"
)

const glRenderVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _v { float verts[]; };
void main() {
	gl_Position = vec4(verts[gl_VertexID * 2], verts[gl_VertexID * 2 + 1], 0.0, 1.0);
}`

const glRenderFrag = `#version 310 es
precision highp float;
out vec4 fragColor;
void main() { fragColor = vec4(0.0, 1.0, 0.0, 1.0); }`

func TestGLBackendRender(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL backend render test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	const w, h = 64, 64
	tex, err := dev.NewTexture(gpu.TextureDescriptor{
		Format: gpu.RGBA8Unorm, Width: w, Height: h, RenderTarget: true,
	})
	if err != nil {
		t.Fatalf("NewTexture: %v", err)
	}
	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: glRenderVert})
	if err != nil {
		t.Fatalf("vertex module: %v", err)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: glRenderFrag})
	if err != nil {
		t.Fatalf("fragment module: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("NewRenderPipeline: %v", err)
	}

	// A triangle that covers the whole viewport.
	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(verts), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: tex, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 1, 1}, // blue background
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	pix := tex.ReadPixels()
	if len(pix) != w*h*4 {
		t.Fatalf("ReadPixels returned %d bytes, want %d", len(pix), w*h*4)
	}
	// The triangle covers the whole viewport, so the center is green (not the
	// blue clear color), proving draw + the FBO render target + readback work.
	c := (h/2*w + w/2) * 4
	if pix[c] != 0 || pix[c+1] != 255 || pix[c+2] != 0 {
		t.Fatalf("center pixel = (%d,%d,%d), want green (0,255,0)", pix[c], pix[c+1], pix[c+2])
	}
}
