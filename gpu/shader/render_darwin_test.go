// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// End-to-end vertex/fragment Go→shader: author a triangle's vertex and fragment
// shaders in Go, compile them to MSL, build a render pipeline, render headless
// through the Device API, and assert the result. cgo-free.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// Go-authored vertex + fragment shaders. Vec4 maps to MSL float4; the //gpu:
// directives select the pipeline stage.
const renderKernels = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

//gpu:vertex
func VTri(vid uint, verts []float32) Vec4 {
	return Vec4{verts[vid*2], verts[vid*2+1], 0, 1}
}

//gpu:fragment
func FRed() Vec4 {
	return Vec4{1, 0, 0, 1}
}
`

func TestGoShaderRenderTriangle(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(renderKernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	if ks["VTri"].Stage != shader.StageVertex || ks["FRed"].Stage != shader.StageFragment {
		t.Fatalf("stages misdetected: VTri=%v FRed=%v", ks["VTri"].Stage, ks["FRed"].Stage)
	}

	vmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VTri"].MSL})
	if err != nil {
		t.Fatalf("vertex MSL compile: %v\n%s", err, ks["VTri"].MSL)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FRed"].MSL})
	if err != nil {
		t.Fatalf("fragment MSL compile: %v\n%s", err, ks["FRed"].MSL)
	}

	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VTri",
		FragmentModule: fmod, FragmentEntry: "FRed",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}

	const W, H = 16, 16
	target, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})
	if err != nil {
		t.Fatalf("texture: %v", err)
	}

	verts := []float32{-1, -1, 3, -1, -1, 3}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: len(verts) * 4, Usage: gpu.BufferStorage, Data: f32bytes(verts)})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{ColorTexture: target, Load: gpu.LoadClear, ClearColor: [4]float64{0, 1, 0, 1}})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, 3)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	px := target.ReadPixels()
	i := (H/2*W + W/2) * 4
	if px[i] < 200 || px[i+1] > 60 {
		t.Fatalf("center pixel = %v, want red (Go-authored shaders)", px[i:i+4])
	}
}

func f32bytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
