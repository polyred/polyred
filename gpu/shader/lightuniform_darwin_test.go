// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Diffuse lighting with a light supplied as a fragment uniform, and vector
// locals (float4 n, l) — proving the compiler's fragment-uniform convention
// (first struct param = stage_in, rest = uniform) and vector-preserving type
// inference for normalize(). Toward the renderer's deferred lighting pass.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

const lightUniformKernels = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos    Vec4 ` + "`gpu:\"position\"`" + `
	Normal Vec4
}

type Light struct {
	Dir Vec4
}

//gpu:vertex
func VLU(vid uint, pos []float32, nrm []float32) VOut {
	return VOut{
		Pos:    Vec4{pos[vid*2], pos[vid*2+1], 0, 1},
		Normal: Vec4{nrm[vid*3], nrm[vid*3+1], nrm[vid*3+2], 0},
	}
}

//gpu:fragment
func FLU(in VOut, light Light) Vec4 {
	n := normalize(in.Normal)
	l := normalize(light.Dir)
	d := max(dot(n, l), 0.0)
	return Vec4{d, d, d, 1}
}
`

func TestGoShaderLightUniform(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(lightUniformKernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	// FLU's bindings: Light is the lone uniform at binding 0 (stage_in is not a binding).
	if len(ks["FLU"].Bindings) != 1 || ks["FLU"].Bindings[0].Kind != shader.UniformBuffer {
		t.Fatalf("FLU bindings = %+v, want one uniform", ks["FLU"].Bindings)
	}

	vmod, _ := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["VLU"].MSL})
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["FLU"].MSL})
	if err != nil {
		t.Fatalf("fragment MSL: %v\n%s", err, ks["FLU"].MSL)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "VLU",
		FragmentModule: fmod, FragmentEntry: "FLU",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	const W, H = 16, 16
	target, _ := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: W, Height: H, RenderTarget: true})

	pos := []float32{-1, -1, 3, -1, -1, 3}
	nrm := []float32{0, 0.6, 0.8, 0, 0.6, 0.8, 0, 0.6, 0.8}
	light := []float32{0, 0, 1, 0} // Light{Dir: +Z}
	posBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(pos) * 4, Usage: gpu.BufferStorage, Data: lub(pos)})
	nrmBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(nrm) * 4, Usage: gpu.BufferStorage, Data: lub(nrm)})
	lightBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 16, Usage: gpu.BufferUniform, Data: lub(light)})

	layout := dev.NewBindGroupLayout(gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageFragment, Kind: gpu.UniformBuffer})
	bg := dev.NewBindGroup(layout, gpu.BindGroupEntry{Binding: 0, Buffer: lightBuf})

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
	lit := int(px[i])
	// N·L = (0,0.6,0.8)·(0,0,1) = 0.8 -> ~204.
	if lit < 188 || lit > 220 {
		t.Fatalf("center diffuse (uniform light) = %d, want ~204", lit)
	}
}

func lub(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
