// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// GPU lighting math: a fragment shader written in Go computes diffuse shading
// (normalize + dot) over an interpolated normal. Proves real per-fragment
// shading math runs on the GPU through the abstraction — the core of the
// renderer's deferred pass — cgo-free.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// Diffuse lighting: d = max(dot(normalize(N), L), 0). L is +Z. With N =
// (0,0.6,0.8) the result is 0.8 → ~204 grayscale.
const lightingKernels = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos    Vec4 ` + "`gpu:\"position\"`" + `
	Normal Vec4
}

//gpu:vertex
func VLit(vid uint, pos []float32, nrm []float32) VOut {
	return VOut{
		Pos:    Vec4{pos[vid*2], pos[vid*2+1], 0, 1},
		Normal: Vec4{nrm[vid*3], nrm[vid*3+1], nrm[vid*3+2], 0},
	}
}

//gpu:fragment
func FLit(in VOut) Vec4 {
	d := max(dot(normalize(in.Normal), Vec4{0, 0, 1, 0}), 0.0)
	return Vec4{d, d, d, 1}
}
`

func TestGoShaderDiffuseLighting(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(lightingKernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	vmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VLit"].MSL})
	if err != nil {
		t.Fatalf("vertex MSL: %v\n%s", err, ks["VLit"].MSL)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FLit"].MSL})
	if err != nil {
		t.Fatalf("fragment MSL: %v\n%s", err, ks["FLit"].MSL)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VLit",
		FragmentModule: fmod, FragmentEntry: "FLit",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	const W, H = 16, 16
	target, _ := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})

	pos := []float32{-1, -1, 3, -1, -1, 3}
	nrm := []float32{0, 0.6, 0.8, 0, 0.6, 0.8, 0, 0.6, 0.8}
	posBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(pos) * 4, Usage: gpu.BufferStorage, Data: lb(pos)})
	nrmBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(nrm) * 4, Usage: gpu.BufferStorage, Data: lb(nrm)})

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{ColorTexture: target, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1}})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, posBuf)
	rp.SetVertexBuffer(1, nrmBuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	px := target.ReadPixels()
	i := (H/2*W + W/2) * 4
	lit := int(px[i])
	// dot((0,0.6,0.8),(0,0,1)) = 0.8 -> 0.8*255 ~= 204.
	if lit < 188 || lit > 220 {
		t.Fatalf("center diffuse = %d, want ~204 (N·L=0.8)", lit)
	}
	if px[i] != px[i+1] || px[i+1] != px[i+2] {
		t.Fatalf("expected grayscale, got (%d,%d,%d)", px[i], px[i+1], px[i+2])
	}
}

func lb(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
