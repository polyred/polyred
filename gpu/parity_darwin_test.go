// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package gpu_test

import (
	"testing"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

// TestShadingParityMetal runs the shared cross-backend shading parity on the
// Metal backend (Go kernel -> MSL). Runs in the macOS CI job.
func TestShadingParityMetal(t *testing.T) {
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverMetal))
	if err != nil {
		t.Skipf("no Metal device: %v", err)
	}
	defer dev.Close()
	runParity(t, dev, func(goSrc, entry string) (*gpu.ShaderModule, []shader.Binding, error) {
		ks, err := shader.Compile(goSrc)
		if err != nil {
			return nil, nil, err
		}
		mod, err := dev.NewShaderModule(gpu.ShaderSource{MSL: ks[entry].MSL})
		if err != nil {
			return nil, nil, err
		}
		return mod, ks[entry].Bindings, nil
	})
}
