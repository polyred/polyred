// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Depth-buffer conformance for the GL backend (forward-rasterizer brick 3a, the
// GL CI counterpart of the Metal-only gpu-render-depth.md): two full-screen
// triangles at different depths are drawn in BOTH orders into a color + depth
// target; with the depth attachment the nearer (red) triangle must win at the
// center regardless of draw order. The reverse order is the discriminating case:
// without depth testing the last-drawn (far, green) triangle would win. Runs in
// CI on Mesa llvmpipe (software, surfaceless).
package gpu_test

import (
	"os"
	"testing"

	"poly.red/gpu"
)

const depthGLVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _pos { float pos[]; };
layout(std430, binding = 1) readonly buffer _col { float col[]; };
out vec3 vcolor;
void main() {
	int i = gl_VertexID;
	gl_Position = vec4(pos[i*3], pos[i*3+1], pos[i*3+2], 1.0);
	vcolor = vec3(col[i*3], col[i*3+1], col[i*3+2]);
}`

const depthGLFrag = `#version 310 es
precision highp float;
in vec3 vcolor;
out vec4 fragColor;
void main() { fragColor = vec4(vcolor, 1.0); }`

func TestGLRenderDepthOcclusion(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL depth test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: depthGLVert})
	if err != nil {
		t.Fatalf("vertex module: %v", err)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: depthGLFrag})
	if err != nil {
		t.Fatalf("fragment module: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat: gpu.RGBA8Unorm,
		DepthFormat: gpu.Depth32Float,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	// GL clip-space z is in [-1,1] (depth = (z+1)/2); with compare "less" a smaller
	// z is nearer. Full-screen triangles (the big-triangle trick) at z=-0.5 (near,
	// red) and z=+0.5 (far, green); both cover the center pixel.
	near := []float32{-1, -1, -0.5, 3, -1, -0.5, -1, 3, -0.5}
	far := []float32{-1, -1, 0.5, 3, -1, 0.5, -1, 3, 0.5}
	red := []float32{1, 0, 0, 1, 0, 0, 1, 0, 0}
	green := []float32{0, 1, 0, 0, 1, 0, 0, 1, 0}

	buf := func(d []float32) *gpu.Buffer {
		b, err := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(d), Usage: gpu.BufferStorage})
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
			ColorTexture: color, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
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

		pix := color.ReadPixels()
		c := ((H/2)*W + W/2) * 4
		return pix[c], pix[c+1], pix[c+2]
	}

	for _, firstFar := range []bool{false, true} {
		r, g, b := render(firstFar)
		if r < 200 || g > 60 {
			t.Fatalf("firstFar=%v: center=(%d,%d,%d), want near (red) to win via depth test", firstFar, r, g, b)
		}
	}
}
