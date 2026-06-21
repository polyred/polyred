// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Conformance test for the presentable swapchain (Surface) API, exercised
// headless through the GL backend: acquire a frame texture, render into it,
// present, and read the presented frame back. Also checks double-buffering
// rotation and resize. The on-screen attachment is the only display-gated piece;
// this verifies everything up to it. Runs in CI on Mesa llvmpipe.
package gpu_test

import (
	"os"
	"testing"

	"poly.red/gpu"
)

func TestSurfaceHeadlessPresent(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the Surface conformance test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	const w, h = 32, 32
	surf, err := dev.CreateSurface(gpu.SurfaceDescriptor{
		Width: w, Height: h, Format: gpu.RGBA8Unorm, Frames: 2,
	})
	if err != nil {
		t.Fatalf("CreateSurface: %v", err)
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
	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(verts), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	var acquired []*gpu.Texture
	for frame := 0; frame < 3; frame++ {
		tex := surf.AcquireNextTexture()
		acquired = append(acquired, tex)

		enc := dev.NewCommandEncoder()
		rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
			ColorTexture: tex, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 1, 1},
		})
		rp.SetPipeline(pipe)
		rp.SetVertexBuffer(0, vbuf)
		rp.Draw(gpu.TriangleList, 0, 3)
		rp.End()
		dev.Queue().Submit(enc.Finish())

		if err := surf.Present(); err != nil {
			t.Fatalf("Present frame %d: %v", frame, err)
		}
		pix := surf.Texture().ReadPixels()
		c := (h/2*w + w/2) * 4
		if pix[c] != 0 || pix[c+1] != 255 || pix[c+2] != 0 {
			t.Fatalf("frame %d center = (%d,%d,%d), want green", frame, pix[c], pix[c+1], pix[c+2])
		}
	}

	// Double buffering: consecutive frames use distinct textures; a 2-frame ring
	// reuses the first texture on the third frame.
	if acquired[0] == acquired[1] {
		t.Errorf("expected distinct textures for frames 0 and 1 (double buffering)")
	}
	if acquired[0] != acquired[2] {
		t.Errorf("expected the 2-frame ring to reuse the frame-0 texture on frame 2")
	}

	// Present without Acquire is an error.
	if err := surf.Present(); err == nil {
		t.Errorf("expected Present without Acquire to error")
	}

	// Resize reallocates and resets the swapchain.
	if err := surf.Resize(16, 16); err != nil {
		t.Fatalf("Resize: %v", err)
	}
	if gw, gh := surf.Size(); gw != 16 || gh != 16 {
		t.Errorf("after resize Size() = (%d,%d), want (16,16)", gw, gh)
	}
	rt := surf.AcquireNextTexture()
	if rt.Width() != 16 || rt.Height() != 16 {
		t.Errorf("resized texture = %dx%d, want 16x16", rt.Width(), rt.Height())
	}
}
