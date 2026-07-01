// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package render

import (
	"testing"

	"poly.red/gpu"
)

// TestGPUForwardMetal verifies the GPU forward rasterizer runs on the Metal backend
// (the darwin runtime, as opposed to GL which is the CI oracle): a full Render() with
// a Metal device must run BOTH the forward and deferred passes on the GPU (not fall
// back to the CPU) and match the all-CPU render within the measured parity tolerance.
// It exercises the MSL forward shaders (fwdGBufMSL) and, in particular, pins the
// Metal front-facing / back-face-cull convention against the CPU: if it were
// inverted the bunny would render inside-out and blow past the tolerance.
func TestGPUForwardMetal(t *testing.T) {
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverMetal))
	if err != nil {
		t.Skipf("no Metal device: %v", err)
	}
	defer dev.Close()

	const w, h = 96, 96
	s, c := newscene(w, h)

	// Coverage gate: the Metal forward pass must cover the same fragments as the CPU
	// (pins the z-remap to Metal's [0,1] clip and the inverted back-face-cull winding;
	// a broken z-remap covers ~0, an inverted cull covers the wrong faces).
	probe := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), GPU(dev))
	pb := probe.CurrBuffer()
	pb.Clear()
	if err := probe.gpuForwardPass(); err != nil {
		t.Fatalf("gpuForwardPass on Metal: %v", err)
	}
	cr := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU())
	cpuG := cr.CurrBuffer()
	cpuG.Clear()
	cr.cpuForwardPass()
	var nOk, nCPU, ovNo int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			m := pb.UnsafeGet(x, y).Ok
			cpuOk := cpuG.UnsafeGet(x, y).Ok
			if m {
				nOk++
			}
			if cpuOk {
				nCPU++
			}
			if m && cpuOk {
				ovNo++
			}
		}
	}
	t.Logf("Metal forward coverage: %d (cpu=%d, overlap=%d)", nOk, nCPU, ovNo)
	if ovNo < nCPU*95/100 || nOk > nCPU*105/100 {
		t.Fatalf("Metal forward coverage off: nOk=%d overlap=%d cpu=%d (want overlap>=95%%, count within 5%%)", nOk, ovNo, nCPU)
	}

	cpu := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU()).Render()

	gr := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), GPU(dev))
	gpuImg := gr.Render()
	if !gr.passOnGPU("forward") {
		t.Fatal("forward pass did not run on Metal (fell back to the CPU)")
	}
	if !gr.passOnGPU("deferred") {
		t.Fatal("deferred pass did not run on Metal (fell back to the CPU)")
	}

	if len(cpu.Pix) != len(gpuImg.Pix) {
		t.Fatalf("size mismatch: cpu %d gpu %d", len(cpu.Pix), len(gpuImg.Pix))
	}
	var n8, n16 int
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > 8 {
			n8++
		}
		if d > 16 {
			n16++
		}
	}
	f8 := float64(n8) / float64(len(cpu.Pix))
	t.Logf("Metal forward+deferred vs all-CPU: %.2f%%@>8 %.2f%%@>16", f8*100, 100*float64(n16)/float64(len(cpu.Pix)))
	// Same boundary parity band as GL (measured 4.38%@>8), with headroom for any
	// Metal-vs-GL rasterizer rounding differences. A blown convention (inside-out)
	// would be tens of percent.
	if f8 > 0.08 {
		t.Fatalf("Metal forward+deferred diverges from CPU on %.2f%%@>8; want <8%%", f8*100)
	}
}
