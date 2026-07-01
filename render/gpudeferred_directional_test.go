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

// directionalScene is newscene but lit by a directional light, to exercise the
// GPU deferred path's directional-light branch.
func directionalScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(
		light.NewDirectional(
			light.Intensity(0.9),
			light.Color(color.RGBA{R: 255, G: 240, B: 220, A: 255}),
			light.Direction(math.NewVec3[float32](-1, -1, -1)),
		),
		light.NewAmbient(light.Intensity(0.3)),
	)
	m := model.MustLoad("../internal/testdata/bunny.obj")
	m.Rotate(math.NewVec3[float32](0, 1, 0), -math.Pi/6)
	m.Scale(4, 4, 4)
	m.Translate(0.1, 0, -0.2)
	s.Add(m)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 1.5, 1)),
		camera.LookAt(math.NewVec3[float32](0, 0, -0.5), math.NewVec3[float32](0, 1, 0)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 3),
	)
}

func TestGPUDeferredDirectional(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 150, 150
	s, c := directionalScene(w, h)
	opts := []Option{Camera(c), Size(w, h), MSAA(1), Scene(s), Background(color.RGBA{R: 0, G: 127, B: 255, A: 255})}

	cpu := NewRenderer(append(opts, CPU())...).Render()
	gr := NewRenderer(append(opts, GPU(dev), forwardOnCPU())...)
	gpuImg := gr.Render()
	if !gr.passOnGPU("deferred") {
		t.Fatal("GPU deferred path not exercised (directional light)")
	}

	maxDiff := 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > maxDiff {
			maxDiff = d
		}
	}
	if maxDiff > 2 {
		t.Fatalf("CPU vs GPU deferred (directional): max channel diff = %d", maxDiff)
	}
	t.Logf("GPU deferred directional-light parity: max channel diff = %d", maxDiff)
}
