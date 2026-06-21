// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Integration test: the renderer's actual deferred Blinn-Phong compute kernel
// (verbatim from render/gpudeferred.go) compiled to GLSL and run on the cgo-free
// GL backend through the Device API, with its Scene struct uniform marshaled in
// std140. The result for a controlled one-light, one-material fragment is checked
// against the hand-computed Blinn-Phong value. This proves the engine's real
// compute workload (loops, multiple storage buffers, a multi-field uniform,
// vector math, pow) runs on the second backend, not just toy kernels. Runs in CI
// on Mesa llvmpipe (software, surfaceless).
package gpu_test

import (
	"encoding/binary"
	"math"
	"os"
	"testing"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// deferredShadeKernel is a verbatim copy of render/gpudeferred.go's deferred
// Blinn-Phong kernel (see gpu/shader/validate_test.go's deferredKernelSrc).
const deferredShadeKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type Scene struct {
	CamPos    Vec4
	AmbientI  float32
	NumLights float32
	Pad1      float32
	Pad2      float32
}

func Shade(gid uint, normals []float32, worldpos []float32, basecol []float32, lights []float32, matidx []float32, materials []float32, s Scene, out []float32) {
	N := Vec4{normals[gid*4], normals[gid*4+1], normals[gid*4+2], normals[gid*4+3]}
	wpos := Vec4{worldpos[gid*4], worldpos[gid*4+1], worldpos[gid*4+2], worldpos[gid*4+3]}
	col := Vec4{basecol[gid*4], basecol[gid*4+1], basecol[gid*4+2], basecol[gid*4+3]}

	mi := int(matidx[gid])
	diffuse := Vec4{materials[mi*9], materials[mi*9+1], materials[mi*9+2], materials[mi*9+3]}
	specular := Vec4{materials[mi*9+4], materials[mi*9+5], materials[mi*9+6], materials[mi*9+7]}
	shininess := materials[mi*9+8]

	acc := col * s.AmbientI
	count := int(s.NumLights)
	for i := 0; i < count; i++ {
		lt := lights[i*10]
		lp := Vec4{lights[i*10+1], lights[i*10+2], lights[i*10+3], lights[i*10+4]}
		lc := Vec4{lights[i*10+5], lights[i*10+6], lights[i*10+7], lights[i*10+8]}
		li := lights[i*10+9]
		var L Vec4
		var I float32
		if lt < 0.5 {
			Ldir := lp - wpos
			L = normalize(Ldir)
			I = li / length(Ldir)
		} else {
			L = Vec4{-lp.X, -lp.Y, -lp.Z, 0}
			I = li
		}
		V := normalize(s.CamPos - wpos)
		H := normalize(L + V)
		Ld := clamp(dot(N, L), 0.0, 1.0)
		Ls := pow(clamp(dot(N, H), 0.0, 1.0), shininess)
		acc = acc + diffuse*(col*(Ld*I))/255.0 + specular*(lc*(Ls*I))/255.0
	}
	out[gid*4] = acc.X
	out[gid*4+1] = acc.Y
	out[gid*4+2] = acc.Z
	out[gid*4+3] = col.W
}
`

func TestGLBackendDeferredKernel(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL backend deferred-kernel test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.CompileGLSL(deferredShadeKernel)
	if err != nil {
		t.Fatalf("CompileGLSL: %v", err)
	}
	k := ks["Shade"]
	mod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: k.GLSL})
	if err != nil {
		t.Fatalf("NewShaderModule: %v", err)
	}

	// One fragment, one point light, one material, chosen so the Blinn-Phong math
	// is exact: N=V=L=H=+Z, so both clamp(dot)=1 and the result is
	//   col*Ambient + diffuse*col*(1*I)/255 + specular*lc*(1*I)/255.
	// With col=100, Ambient=0.5, I=255, diffuse=specular=lc=1: 50 + 100 + 1 = 151.
	normals := []float32{0, 0, 1, 0}
	worldpos := []float32{0, 0, 0, 0}
	basecol := []float32{100, 100, 100, 255}
	matidx := []float32{0}
	materials := []float32{1, 1, 1, 1, 1, 1, 1, 1, 32}  // diffuse, specular, shininess
	lights := []float32{0, 0, 0, 1, 0, 1, 1, 1, 0, 255} // point: lt,lp(xyzw),lc(xyzw),li
	out := make([]float32, 4)

	// Scene as std140: vec4 CamPos at 0; floats AmbientI/NumLights/Pad1/Pad2 at
	// 16/20/24/28; block size 32.
	scene := make([]byte, 32)
	putF := func(off int, v float32) { binary.LittleEndian.PutUint32(scene[off:], math.Float32bits(v)) }
	putF(8, 1)    // CamPos.Z = 1
	putF(16, 0.5) // AmbientI
	putF(20, 1)   // NumLights

	storage := map[string][]float32{
		"normals": normals, "worldpos": worldpos, "basecol": basecol,
		"lights": lights, "matidx": matidx, "materials": materials, "out": out,
	}

	var le []gpu.BindGroupLayoutEntry
	var bge []gpu.BindGroupEntry
	for _, bd := range k.Bindings {
		kind := gpu.StorageBuffer
		var gb *gpu.Buffer
		if bd.Kind == shader.UniformBuffer {
			kind = gpu.UniformBuffer
			gb, _ = dev.NewBuffer(gpu.BufferDescriptor{Data: scene, Usage: gpu.BufferUniform})
		} else {
			gb, _ = dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(storage[bd.Name]), Usage: gpu.BufferStorage})
		}
		le = append(le, gpu.BindGroupLayoutEntry{Binding: bd.Index, Visibility: gpu.StageCompute, Kind: kind})
		bge = append(bge, gpu.BindGroupEntry{Binding: bd.Index, Buffer: gb})
	}
	layout := dev.NewBindGroupLayout(le...)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shade",
	})
	if err != nil {
		t.Fatalf("NewComputePipeline: %v", err)
	}
	bg := dev.NewBindGroup(layout, bge...)

	// Find the output buffer to read back.
	var outBuf *gpu.Buffer
	for i, bd := range k.Bindings {
		if bd.Name == "out" {
			outBuf = bge[i].Buffer
		}
	}

	enc := dev.NewCommandEncoder()
	pass := enc.BeginComputePass()
	pass.SetPipeline(pipe)
	pass.SetBindGroup(0, bg)
	pass.Dispatch(1, 1, 1)
	pass.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := glFloatsOf(outBuf.Bytes(), 4)
	want := []float32{151, 151, 151, 255}
	for i := range want {
		if d := got[i] - want[i]; d > 1e-2 || d < -1e-2 {
			t.Fatalf("deferred Shade out = %v, want %v", got, want)
		}
	}
	t.Logf("deferred Blinn-Phong kernel on GL: out = %v (matches CPU)", got)
}
