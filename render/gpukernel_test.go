// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"testing"

	"poly.red/gpu"
	"poly.red/gpu/shader/gpumath/kernels"
)

// TestKernelSourceBackend verifies render selects the right shading language per
// device backend: MSL for Metal, GLSL for GL, unsupported elsewhere. Device-free
// (shader.Compile/CompileGLSL are pure Go), so it runs in standard CI on every
// platform without opening a GPU.
func TestKernelSourceBackend(t *testing.T) {
	metal, err := kernelSource(gpu.DriverMetal, kernels.ShadeSrc, "Shade")
	if err != nil {
		t.Fatalf("Metal: %v", err)
	}
	if metal.MSL == "" || metal.GLSL != "" {
		t.Errorf("Metal: want MSL only, got MSL=%d GLSL=%d bytes", len(metal.MSL), len(metal.GLSL))
	}

	gl, err := kernelSource(gpu.DriverGL, kernels.ShadeSrc, "Shade")
	if err != nil {
		t.Fatalf("GL: %v", err)
	}
	if gl.GLSL == "" || gl.MSL != "" {
		t.Errorf("GL: want GLSL only, got MSL=%d GLSL=%d bytes", len(gl.MSL), len(gl.GLSL))
	}

	if _, err := kernelSource(gpu.DriverVulkan, kernels.ShadeSrc, "Shade"); err != errKernelBackendUnsupported {
		t.Errorf("Vulkan: want errKernelBackendUnsupported, got %v", err)
	}
}
