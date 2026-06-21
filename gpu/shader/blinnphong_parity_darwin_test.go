// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Blinn-Phong deferred-shading parity: the engine's shader.FragmentShader
// (shader/blinn_old.go) re-authored as a Go GPU kernel, run through the Device
// API, and checked numerically against the engine for a constructed
// material + point light + ambient + fragment. This proves the deferred
// shading model — the heart of passDeferred — computes identically on the GPU.
package shader_test

import (
	stdmath "math"
	"testing"
	"unsafe"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/gpu"
	gpushader "poly.red/gpu/shader"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	eshader "poly.red/shader"
)

// blinnKernel re-expresses shader.FragmentShader's single-point-light + ambient
// path (white base texture) in Go. Params are packed into one float buffer.
const blinnKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

func Shade(gid uint, p []float32, out []float32) {
	N := Vec4{p[0], p[1], p[2], p[3]}
	wpos := Vec4{p[4], p[5], p[6], p[7]}
	col := Vec4{p[8], p[9], p[10], p[11]}
	camPos := Vec4{p[12], p[13], p[14], p[15]}
	lightPos := Vec4{p[16], p[17], p[18], p[19]}
	lightColor := Vec4{p[20], p[21], p[22], p[23]}
	diffuse := Vec4{p[24], p[25], p[26], p[27]}
	specular := Vec4{p[28], p[29], p[30], p[31]}
	lightI := p[32]
	ambI := p[33]
	shininess := p[34]

	La := col * ambI
	Ldir := lightPos - wpos
	L := normalize(Ldir)
	I := lightI / length(Ldir)
	V := normalize(camPos - wpos)
	H := normalize(L + V)
	Ld := clamp(dot(N, L), 0.0, 1.0)
	Ls := pow(clamp(dot(N, H), 0.0, 1.0), shininess)
	LdC := col * (Ld * I)
	LsC := lightColor * (Ls * I)
	final := La + diffuse*LdC/255.0 + specular*LsC/255.0
	out[0] = final.X
	out[1] = final.Y
	out[2] = final.Z
}
`

func TestBlinnPhongDeferredParity(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	// Construct the engine inputs.
	mat := material.NewBlinnPhong(
		material.Texture(buffer.NewTexture()), // default 1x1 white
		material.Diffuse(color.FromValue(0.6, 0.6, 0.6, 1.0)),
		material.Specular(color.FromValue(0.7, 0.7, 0.7, 1.0)),
		material.Shininess(25),
	)

	n := math.NewVec3[float32](0, 0, 1).ToVec4(0).Unit()
	x := math.NewVec3[float32](0.1, 0.2, 0.0).ToVec4(1)
	frag := primitive.Fragment{U: 1, V: 1, Du: 1, Dv: 1, Nor: n, FaceNor: n, WordPos: x}
	info := buffer.Fragment{Ok: true, Fragment: frag}

	camPos := math.NewVec3[float32](0, 0, 5)
	lpos := math.NewVec3[float32](2, 3, 4)
	pt := light.NewPoint(light.Intensity(1), light.Color(color.RGBA{R: 200, G: 150, B: 100, A: 255}), light.Position(lpos))
	amb := light.NewAmbient(light.Intensity(0.1), light.Color(color.RGBA{R: 255, G: 255, B: 255, A: 255}))

	want := eshader.FragmentShader(mat, info, camPos, []light.Source{pt}, []light.Environment{amb})

	// Marshal the same inputs for the GPU kernel. Base color is the 1x1 white
	// texture; ambient color is unused by the shader (only its intensity).
	p := []float32{
		n.X, n.Y, n.Z, 0, // N
		x.X, x.Y, x.Z, 1, // world pos
		255, 255, 255, 255, // base color (white texture)
		camPos.X, camPos.Y, camPos.Z, 1, // cam pos
		lpos.X, lpos.Y, lpos.Z, 1, // light pos
		200, 150, 100, 255, // light color
		float32(mat.Diffuse.R), float32(mat.Diffuse.G), float32(mat.Diffuse.B), float32(mat.Diffuse.A),
		float32(mat.Specular.R), float32(mat.Specular.G), float32(mat.Specular.B), float32(mat.Specular.A),
		1,   // light intensity
		0.1, // ambient intensity
		float32(mat.Shininess),
	}

	got := runShade(t, dev, p)

	for i, ch := range []uint8{want.R, want.G, want.B} {
		g := uint8(math.Clamp(float32(stdmath.Round(float64(got[i]))), 0, 255))
		if d := int(g) - int(ch); d < -1 || d > 1 {
			t.Fatalf("channel %d: GPU Blinn-Phong = %d, engine = %d (diff %d)", i, g, ch, d)
		}
	}
	t.Logf("Blinn-Phong deferred parity OK: GPU≈engine RGB=(%d,%d,%d)", want.R, want.G, want.B)
}

func runShade(t *testing.T, dev *gpu.Device, params []float32) []float32 {
	t.Helper()
	ks, err := gpushader.Compile(blinnKernel)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["Shade"].MSL})
	if err != nil {
		t.Fatalf("MSL: %v\n%s", err, ks["Shade"].MSL)
	}
	layout := dev.NewBindGroupLayout(
		gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
		gpu.BindGroupLayoutEntry{Binding: 1, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shade"})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	inBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: len(params) * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: pbytes(params)})
	outBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 4 * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	bg := dev.NewBindGroup(layout, gpu.BindGroupEntry{Binding: 0, Buffer: inBuf}, gpu.BindGroupEntry{Binding: 1, Buffer: outBuf})
	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(1, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	out := make([]float32, 4)
	copy(out, unsafe.Slice((*float32)(unsafe.Pointer(&outBuf.Bytes()[0])), 4))
	return out
}

func pbytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
