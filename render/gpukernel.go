// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"errors"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// errKernelBackendUnsupported signals the device's driver has no render kernel
// compilation path yet (Vulkan needs runtime SPIR-V via glslang, DX12 is
// unimplemented), so the caller's runPass falls back to the CPU.
var errKernelBackendUnsupported = errors.New("render: GPU kernel not supported on this backend")

// kernelSource compiles the Go-DSL kernel src for the given backend driver and
// returns the shading-language ShaderSource for entry: MSL for Metal, GLSL for
// GL. It is the single place render selects a shading language, and is
// device-free (shader.Compile/CompileGLSL are pure Go) so it can be unit-tested
// without a GPU.
func kernelSource(driver gpu.Driver, src, entry string) (gpu.ShaderSource, error) {
	switch driver {
	case gpu.DriverMetal:
		ks, err := shader.Compile(src)
		if err != nil {
			return gpu.ShaderSource{}, err
		}
		return gpu.ShaderSource{MSL: ks[entry].MSL}, nil
	case gpu.DriverGL:
		ks, err := shader.CompileGLSL(src)
		if err != nil {
			return gpu.ShaderSource{}, err
		}
		return gpu.ShaderSource{GLSL: ks[entry].GLSL}, nil
	default:
		return gpu.ShaderSource{}, errKernelBackendUnsupported
	}
}

// kernelModule compiles src for dev's backend and returns a shader module for
// entry. Every render GPU pass goes through here, so the passes are
// backend-agnostic: the same author-once kernel runs on Metal and GL.
func kernelModule(dev *gpu.Device, src, entry string) (*gpu.ShaderModule, error) {
	source, err := kernelSource(dev.Driver(), src, entry)
	if err != nil {
		return nil, err
	}
	return dev.NewShaderModule(source)
}
