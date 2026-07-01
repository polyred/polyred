// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build linux

package render

import (
	"os"
	"testing"

	"poly.red/gpu"
)

// TestGLDeferredRender isolates deferred-shading parity on the cgo-free GL backend:
// it runs the CPU forward pass (so the G-buffer input is identical to the CPU
// reference) then the GPU deferred pass, which routes through the same
// runDeferredKernel that serves Metal, now compiled to GLSL by kernelModule. This
// keeps the gate tight (<2% @>8) as a pure deferred-shading gate -- the full GPU
// pipeline (GPU forward too) is gated separately by TestGPUForwardDeferredIntegration,
// where the forward rasterizer's boundary parity band lives. It asserts the deferred
// pass actually ran on the GPU. Runs in CI on Mesa llvmpipe (surfaceless); skipped
// under standard `go test ./...` (env unset) so it never opens GL on a runner without
// surfaceless Mesa (which would segfault, not error).
func TestGLDeferredRender(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GL render test")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()
	if dev.Driver() != gpu.DriverGL {
		t.Fatalf("want DriverGL, got %v", dev.Driver())
	}

	const w, h = 96, 96
	s, c := newscene(w, h)
	// Single worker so the forward pass is deterministic, letting the CPU and GL
	// deferred shading be compared on an identical G-buffer.
	opts := []Option{Scene(s), Camera(c), Size(w, h), Workers(1), BatchSize(1)}

	cpu := NewRenderer(append(opts, CPU())...).Render()

	// Force the CPU forward pass on the GPU renderer so the deferred pass shades the
	// SAME G-buffer as the CPU reference; only the deferred stage differs (GPU vs CPU).
	gr := NewRenderer(append(opts, GPU(dev))...)
	buf := gr.CurrBuffer()
	buf.Clear()
	gr.cpuForwardPass()
	buf.ClearColor()
	gr.passDeferred()
	gr.passAntialiasing()
	gl := gr.outBuf
	if !gr.passOnGPU("deferred") {
		t.Fatal("deferred pass did not run on the GL GPU (fell back to CPU)")
	}

	if len(cpu.Pix) != len(gl.Pix) {
		t.Fatalf("size mismatch: cpu %d gl %d", len(cpu.Pix), len(gl.Pix))
	}
	nBig := 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gl.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > 8 {
			nBig++
		}
	}
	if frac := float64(nBig) / float64(len(cpu.Pix)); frac > 0.02 {
		t.Fatalf("GL vs CPU deferred: %.2f%% of channels differ by >8 (want <2%%, %d/%d)", frac*100, nBig, len(cpu.Pix))
	}
	t.Logf("GL deferred render: %d/%d channels differ by >8", nBig, len(cpu.Pix))
}
