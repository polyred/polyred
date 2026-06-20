// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Go→shader varyings: a vertex shader outputs a struct with [[position]] plus a
// per-vertex color; the rasterizer interpolates it; the fragment shader reads
// the interpolated color via [[stage_in]]. Proves vertex→fragment data flow —
// the prerequisite for real shading — end to end, cgo-free.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// Per-vertex color, interpolated to the fragment. The gpu:"position" tag marks
// the clip-space output; Color is a varying.
const varyingKernels = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos   Vec4 ` + "`gpu:\"position\"`" + `
	Color Vec4
}

//gpu:vertex
func VCol(vid uint, pos []float32, col []float32) VOut {
	return VOut{
		Pos:   Vec4{pos[vid*2], pos[vid*2+1], 0, 1},
		Color: Vec4{col[vid*3], col[vid*3+1], col[vid*3+2], 1},
	}
}

//gpu:fragment
func FCol(in VOut) Vec4 {
	return in.Color
}
`

func TestGoShaderVaryingInterpolation(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(varyingKernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	vmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VCol"].MSL})
	if err != nil {
		t.Fatalf("vertex MSL: %v\n%s", err, ks["VCol"].MSL)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FCol"].MSL})
	if err != nil {
		t.Fatalf("fragment MSL: %v\n%s", err, ks["FCol"].MSL)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VCol",
		FragmentModule: fmod, FragmentEntry: "FCol",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	const W, H = 16, 16
	target, _ := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})

	// Full-clip triangle; vertices coloured red, green, blue.
	pos := []float32{-1, -1, 3, -1, -1, 3}
	col := []float32{1, 0, 0, 0, 1, 0, 0, 0, 1}
	posBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(pos) * 4, Usage: gpu.BufferStorage, Data: fb(pos)})
	colBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(col) * 4, Usage: gpu.BufferStorage, Data: fb(col)})

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{ColorTexture: target, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1}})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, posBuf)
	rp.SetVertexBuffer(1, colBuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	px := target.ReadPixels()
	i := (H/2*W + W/2) * 4
	r, g, b := int(px[i]), int(px[i+1]), int(px[i+2])

	// At the center the barycentric weights are (0.5, 0.25, 0.25), so the
	// interpolated colour is ~ (128, 64, 64). Allow generous tolerance for
	// pixel-center sampling.
	near := func(got, want int) bool { return got >= want-30 && got <= want+30 }
	if !near(r, 128) || !near(g, 64) || !near(b, 64) {
		t.Fatalf("center colour = (%d,%d,%d), want ~(128,64,64) from interpolated varyings", r, g, b)
	}
}

func fb(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
