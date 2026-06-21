// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// Conformance test for the Vulkan backend driven through the public Device API:
// gpu.Open(WithDriver(DriverVulkan)) runs a compute kernel (GLSL compiled to
// SPIR-V by glslang) and the result is checked against the CPU. This proves
// Vulkan is a first-class driver behind the same surface as Metal and GL.
// Verified in CI on Mesa lavapipe; gated on POLYRED_VK_PROBE=1.
package gpu_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"poly.red/gpu"
)

func glslToSPIRV(t *testing.T, src string) []byte {
	t.Helper()
	glslang, err := exec.LookPath("glslangValidator")
	if err != nil {
		t.Skipf("glslangValidator not found: %v", err)
	}
	dir := t.TempDir()
	comp := filepath.Join(dir, "k.comp")
	spv := filepath.Join(dir, "k.spv")
	if err := os.WriteFile(comp, []byte(src), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command(glslang, "-V", "--target-env", "vulkan1.0", comp, "-o", spv).CombinedOutput(); err != nil {
		t.Fatalf("glslang failed: %v\n%s", err, out)
	}
	b, err := os.ReadFile(spv)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestVulkanBackendCompute(t *testing.T) {
	if os.Getenv("POLYRED_VK_PROBE") != "1" {
		t.Skip("set POLYRED_VK_PROBE=1 to run the Vulkan backend conformance test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverVulkan))
	if err != nil {
		t.Skipf("no Vulkan device: %v", err)
	}
	defer dev.Close()
	if dev.Driver() != gpu.DriverVulkan {
		t.Fatalf("expected DriverVulkan, got %v", dev.Driver())
	}

	spv := glslToSPIRV(t, `#version 450
layout(local_size_x = 1) in;
layout(std430, binding = 0) readonly buffer A { float a[]; };
layout(std430, binding = 1) readonly buffer B { float b[]; };
layout(std430, binding = 2) buffer O { float o[]; };
void main() { uint i = gl_GlobalInvocationID.x; o[i] = a[i] + b[i]; }`)

	mod, err := dev.NewShaderModule(gpu.ShaderSource{SPIRV: spv})
	if err != nil {
		t.Fatalf("NewShaderModule: %v", err)
	}
	sb := func(i int) gpu.BindGroupLayoutEntry {
		return gpu.BindGroupLayoutEntry{Binding: i, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer}
	}
	layout := dev.NewBindGroupLayout(sb(0), sb(1), sb(2))
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "main",
	})
	if err != nil {
		t.Fatalf("NewComputePipeline: %v", err)
	}

	const n = 1024
	a := make([]float32, n)
	b := make([]float32, n)
	for i := range a {
		a[i] = float32(i)
		b[i] = float32(2*i + 1)
	}
	aBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(a), Usage: gpu.BufferStorage})
	bBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Data: glBytesOf(b), Usage: gpu.BufferStorage})
	oBuf, _ := dev.NewBuffer(gpu.BufferDescriptor{Size: n * 4, Usage: gpu.BufferStorage})
	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: aBuf},
		gpu.BindGroupEntry{Binding: 1, Buffer: bBuf},
		gpu.BindGroupEntry{Binding: 2, Buffer: oBuf},
	)

	enc := dev.NewCommandEncoder()
	pass := enc.BeginComputePass()
	pass.SetPipeline(pipe)
	pass.SetBindGroup(0, bg)
	pass.Dispatch(n, 1, 1)
	pass.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := glFloatsOf(oBuf.Bytes(), n)
	for i := range got {
		if got[i] != a[i]+b[i] {
			t.Fatalf("Vulkan backend add[%d] = %v, want %v", i, got[i], a[i]+b[i])
		}
	}
	t.Logf("Vulkan backend through the Device API: %d/%d add results match the CPU", n, n)
}
