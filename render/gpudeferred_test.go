// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

// Full passDeferred GPU offload: render a real scene with deferred Blinn-Phong
// shading on the CPU vs on the GPU (via render.GPU(dev)) and assert the images
// match within rounding tolerance. The renderer's deferred shading pass now
// runs on the poly.red/gpu abstraction. cgo-free.
package render

import (
	"image/color"
	"testing"

	"poly.red/gpu"
)

func TestGPUDeferredParity(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 150, 150
	s, c := newscene(w, h)
	opts := []Option{
		Camera(c), Size(w, h), MSAA(1), Scene(s),
		Background(color.RGBA{R: 0, G: 127, B: 255, A: 255}),
	}

	cpu := NewRenderer(append(opts, CPU())...).Render()

	gr := NewRenderer(append(opts, GPU(dev), forwardOnCPU())...)
	gpuImg := gr.Render()
	if !gr.passOnGPU("deferred") {
		t.Fatal("GPU deferred path was not exercised (fell back to CPU)")
	}

	if cpu.Bounds() != gpuImg.Bounds() {
		t.Fatalf("bounds differ: %v vs %v", cpu.Bounds(), gpuImg.Bounds())
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
	if maxDiff > 2 {
		t.Fatalf("CPU vs GPU deferred shading: max channel diff = %d (want <= 2), differing bytes = %d/%d", maxDiff, nDiff, len(cpu.Pix))
	}
	t.Logf("GPU deferred-pass parity: max channel diff = %d over %d/%d bytes", maxDiff, nDiff, len(cpu.Pix))
}
