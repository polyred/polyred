// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Blinn-Phong specular highlight authored in Go: the half-vector term
// pow(clamp(dot(N, H), 0, 1), shininess) with H = normalize(L + V). Together
// with the diffuse test this proves both Blinn-Phong lighting components run on
// the GPU through the abstraction (shader/blinn_old.go), cgo-free.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

const specularKernels = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos    Vec4 ` + "`gpu:\"position\"`" + `
	Normal Vec4
}

type Params struct {
	LightDir Vec4
	ViewDir  Vec4
}

//gpu:vertex
func VSpec(vid uint, pos []float32, nrm []float32) VOut {
	return VOut{
		Pos:    Vec4{pos[vid*2], pos[vid*2+1], 0, 1},
		Normal: Vec4{nrm[vid*3], nrm[vid*3+1], nrm[vid*3+2], 0},
	}
}

//gpu:fragment
func FSpec(in VOut, p Params) Vec4 {
	n := normalize(in.Normal)
	h := normalize(p.LightDir + p.ViewDir)
	s := pow(clamp(dot(n, h), 0.0, 1.0), 32.0)
	return Vec4{s, s, s, 1}
}
`

func TestGoShaderSpecular(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(specularKernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	vmod, _ := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VSpec"].MSL})
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FSpec"].MSL})
	if err != nil {
		t.Fatalf("fragment MSL: %v\n%s", err, ks["FSpec"].MSL)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VSpec",
		FragmentModule: fmod, FragmentEntry: "FSpec",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	const W, H = 16, 16
	target, _ := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})

	pos := []float32{-1, -1, 3, -1, -1, 3}
	nrm := []float32{0, 0, 1, 0, 0, 1, 0, 0, 1} // facing +Z
	// Light and view both along +Z, so H = +Z and N·H = 1 -> full highlight.
	params := []float32{0, 0, 1, 0, 0, 0, 1, 0}
	posBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(pos) * 4, Usage: gpu.BufferStorage, Data: sb(pos)})
	nrmBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(nrm) * 4, Usage: gpu.BufferStorage, Data: sb(nrm)})
	parBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(params) * 4, Usage: gpu.BufferUniform, Data: sb(params)})

	layout := dev.NewBindGroupLayout(gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageFragment, Kind: gpu.UniformBuffer})
	bg := dev.NewBindGroup(layout, gpu.BindGroupEntry{Binding: 0, Buffer: parBuf})

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{ColorTexture: target, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1}})
	rp.SetPipeline(pipe)
	rp.SetBindGroup(0, bg)
	rp.SetVertexBuffer(0, posBuf)
	rp.SetVertexBuffer(1, nrmBuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	px := target.ReadPixels()
	i := (H/2*W + W/2) * 4
	// N·H = 1 -> pow(1, 32) = 1 -> full white highlight.
	if px[i] < 240 {
		t.Fatalf("center specular = %d, want ~255 (N·H=1)", px[i])
	}
}

func sb(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
