// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package render

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/gpu"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
)

func TestGPUDeferredMultiMaterial(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 160, 160
	s := scene.NewScene(
		light.NewPoint(light.Intensity(3), light.Color(color.RGBA{R: 0, G: 0, B: 0, A: 255}), light.Position(math.NewVec3[float32](-2, 2.5, 6))),
		light.NewAmbient(light.Intensity(0.5)),
	)
	bunny := model.MustLoad("../internal/testdata/bunny.obj")
	bunny.Scale(4, 4, 4)
	bunny.Translate(-0.3, 0, -0.2)
	s.Add(bunny)
	gopher := model.MustLoad("../internal/testdata/gopher.obj")
	gopher.Scale(4, 4, 4)
	gopher.Translate(0.4, 0, -0.2)
	s.Add(gopher)

	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 1.5, 1)),
		camera.LookAt(math.NewVec3[float32](0, 0, -0.5), math.NewVec3[float32](0, 1, 0)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 3),
	)

	// Single-worker so the forward pass is deterministic (overlapping objects
	// make the concurrent pass non-deterministic), letting us compare CPU vs
	// GPU shading on an identical G-buffer.
	opts := []Option{Camera(cam), Size(w, h), MSAA(1), Scene(s), Background(color.RGBA{R: 0, G: 127, B: 255, A: 255}), Workers(1), BatchSize(1)}

	cpu := NewRenderer(append(opts, CPU())...).Render()

	debugDeferredSelfCheck = true
	defer func() { debugDeferredSelfCheck = false }()
	gr := NewRenderer(append(opts, GPU(dev))...)
	gpuImg := gr.Render()
	if !gr.passOnGPU("deferred") {
		t.Fatal("GPU deferred path not exercised (multi-material)")
	}

	assertDeferredClose(t, cpu.Pix, gpuImg.Pix, "multi-material")
}
