// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

// Renderer integration: render a real scene with gamma correction on the CPU
// vs on the GPU (via render.GPU(dev)) and assert the images match within the
// LUT-vs-analytic rounding tolerance. Proves the renderer actually uses the
// poly.red/gpu abstraction for a real pass. cgo-free.
package render

import (
	"image/color"
	"testing"

	"poly.red/gpu"
)

func TestGPUGammaParity(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 200, 200
	s, c := newscene(w, h)

	opts := []Option{
		Camera(c), Size(w, h), MSAA(2), Scene(s),
		Background(color.RGBA{R: 0, G: 127, B: 255, A: 255}),
		GammaCorrection(true),
	}

	cpu := NewRenderer(append(opts, CPU())...).Render()
	gpuImg := NewRenderer(append(opts, GPU(dev), forwardOnCPU())...).Render()

	if cpu.Bounds() != gpuImg.Bounds() {
		t.Fatalf("bounds differ: cpu %v gpu %v", cpu.Bounds(), gpuImg.Bounds())
	}

	maxDiff, nDiff := 0, 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > 0 {
			nDiff++
		}
		if d > maxDiff {
			maxDiff = d
		}
	}
	// The CPU path uses a LUT approximation of the analytic sRGB curve the GPU
	// kernel computes, so allow a small per-channel rounding tolerance.
	if maxDiff > 2 {
		t.Fatalf("CPU vs GPU gamma: max channel diff = %d (want <= 2), differing bytes = %d", maxDiff, nDiff)
	}
	t.Logf("renderer GPU-gamma parity: max channel diff = %d over %d differing bytes", maxDiff, nDiff)
}
