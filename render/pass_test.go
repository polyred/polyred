// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"errors"
	"testing"

	"poly.red/gpu"
)

// TestRunPassNoDevice: with no GPU device, runPass runs the CPU closure and
// records the CPU path.
func TestRunPassNoDevice(t *testing.T) {
	r := NewRenderer(CPU()) // force no device
	ran := ""
	r.runPass("x", func() error { ran = "gpu"; return nil }, func() { ran = "cpu" })
	if ran != "cpu" {
		t.Errorf("no device: ran %q, want cpu", ran)
	}
	if r.passOnGPU("x") {
		t.Errorf("no device: passOnGPU should be false")
	}
}

// TestRunPassWithDevice: with a device, a failing GPU closure falls back to CPU
// (records CPU); a succeeding one records GPU. Skips when no device is available.
func TestRunPassWithDevice(t *testing.T) {
	// The render GPU offload is Metal-only today; request Metal so this does not
	// touch a broken headless GL on CI (which segfaults).
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverMetal))
	if err != nil {
		t.Skipf("no Metal device: %v", err)
	}
	defer dev.Close()
	r := NewRenderer(GPU(dev))

	ran := ""
	r.runPass("err", func() error { return errors.New("boom") }, func() { ran = "cpu" })
	if ran != "cpu" || r.passOnGPU("err") {
		t.Errorf("GPU error should fall back to CPU (ran=%q, onGPU=%v)", ran, r.passOnGPU("err"))
	}

	ran = ""
	r.runPass("ok", func() error { ran = "gpu"; return nil }, func() { ran = "cpu" })
	if ran != "gpu" || !r.passOnGPU("ok") {
		t.Errorf("GPU success should record GPU (ran=%q, onGPU=%v)", ran, r.passOnGPU("ok"))
	}
}

// TestGPUByDefault: without an explicit device, NewRenderer acquires one when
// available and the deferred pass runs on the GPU; CPU() forces the CPU path.
func TestGPUByDefault(t *testing.T) {
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverMetal))
	if err != nil {
		t.Skipf("no Metal device (render GPU offload is Metal-only): %v", err)
	}
	dev.Close()
	s, c := newscene(64, 64)
	NewRenderer(Scene(s), Camera(c), Size(64, 64)).Render() // auto-acquire path

	auto := NewRenderer(Scene(s), Camera(c), Size(64, 64))
	auto.Render()
	if !auto.passOnGPU("deferred") {
		t.Errorf("GPU-by-default: deferred should run on GPU")
	}
	forced := NewRenderer(Scene(s), Camera(c), Size(64, 64), CPU())
	forced.Render()
	if forced.passOnGPU("deferred") {
		t.Errorf("CPU(): deferred should run on CPU")
	}
}
