// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

package gpu_test

import (
	"os"
	"strings"
	"testing"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// TestShadingParityGL runs the shared cross-backend shading parity on the GL
// backend (Go kernel -> GLSL). Runs in the gl-probe CI job.
func TestShadingParityGL(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL shading parity")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()
	runShadingParity(t, dev, func(goSrc, entry string) (*gpu.ShaderModule, []shader.Binding, error) {
		ks, err := shader.CompileGLSL(goSrc)
		if err != nil {
			return nil, nil, err
		}
		mod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: ks[entry].GLSL})
		if err != nil {
			return nil, nil, err
		}
		return mod, ks[entry].Bindings, nil
	})
}

// TestShadingParityVulkan runs the shared cross-backend shading parity on the
// Vulkan backend (Go kernel -> GLSL -> SPIR-V via glslang). Runs in the vk-probe
// CI job.
func TestShadingParityVulkan(t *testing.T) {
	if os.Getenv("POLYRED_VK_PROBE") != "1" {
		t.Skip("set POLYRED_VK_PROBE=1 to run the Vulkan shading parity")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverVulkan))
	if err != nil {
		t.Skipf("no Vulkan device: %v", err)
	}
	defer dev.Close()
	runShadingParity(t, dev, func(goSrc, entry string) (*gpu.ShaderModule, []shader.Binding, error) {
		ks, err := shader.CompileGLSL(goSrc)
		if err != nil {
			return nil, nil, err
		}
		// CompileGLSL targets GLES (#version 310 es) for the GL backend; glslang
		// compiles desktop/Vulkan GLSL (#version 450) for SPIR-V. The body is
		// identical, so retarget the version header for the Vulkan path.
		glsl := strings.Replace(ks[entry].GLSL, "#version 310 es", "#version 450", 1)
		spv := glslToSPIRV(t, glsl)
		mod, err := dev.NewShaderModule(gpu.ShaderSource{SPIRV: spv})
		if err != nil {
			return nil, nil, err
		}
		return mod, ks[entry].Bindings, nil
	})
}
