// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

// Mat4 in Go shaders: a kernel multiplies a uniform matrix by a vector and the
// result matches math.Mat4.MulV. Matrices map to MSL float4x4 (column-major);
// this is the piece needed for light-space transforms in shadow mapping.
package shader_test

import (
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
	"poly.red/math"
)

const matrixKernel = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }
type U struct{ M Mat4 }

func MV(gid uint, v []float32, u U, out []float32) {
	vec := Vec4{v[0], v[1], v[2], v[3]}
	r := u.M * vec
	out[0] = r.X
	out[1] = r.Y
	out[2] = r.Z
	out[3] = r.W
}
`

func TestGoShaderMat4(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	ks, err := shader.Compile(matrixKernel)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks["MV"].MSL})
	if err != nil {
		t.Fatalf("MSL: %v\n%s", err, ks["MV"].MSL)
	}
	layout := dev.NewBindGroupLayout(
		gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
		gpu.BindGroupLayoutEntry{Binding: 1, Visibility: gpu.StageCompute, Kind: gpu.UniformBuffer},
		gpu.BindGroupLayoutEntry{Binding: 2, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "MV"})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	m := math.NewMat4[float32](
		2, 0, 0, 5,
		0, 3, 0, 6,
		0, 0, 4, 7,
		0, 0, 0, 1,
	)
	vec := math.NewVec4[float32](1, 2, 3, 1)
	want := m.MulV(vec)

	// MSL float4x4 is column-major: column j, row i = m.Get(i, j).
	mat := make([]float32, 16)
	for j := 0; j < 4; j++ {
		for i := 0; i < 4; i++ {
			mat[j*4+i] = m.Get(i, j)
		}
	}
	v := []float32{vec.X, vec.Y, vec.Z, vec.W}

	vb, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 16, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: matBytes(v)})
	ub, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 64, Usage: gpu.BufferUniform, Data: matBytes(mat)})
	out, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: 16, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: vb},
		gpu.BindGroupEntry{Binding: 1, Buffer: ub},
		gpu.BindGroupEntry{Binding: 2, Buffer: out})

	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(1, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := make([]float32, 4)
	copy(got, unsafe.Slice((*float32)(unsafe.Pointer(&out.Bytes()[0])), 4))
	exp := []float32{want.X, want.Y, want.Z, want.W}
	for i := range exp {
		if d := got[i] - exp[i]; d > 1e-3 || d < -1e-3 {
			t.Fatalf("component %d: GPU=%v want=%v (M·v)", i, got[i], exp[i])
		}
	}
	t.Logf("GPU Mat4·Vec4 = %v matches math.Mat4.MulV", got)
}

func matBytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
