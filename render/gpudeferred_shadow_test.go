// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package render

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/gpu"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
)

// TestGPUDeferredShadow renders a shadow-mapped scene (one shadow-casting light,
// ReceiveShadow materials) on the CPU vs the GPU deferred path and asserts the
// images match. The GPU path applies the shadow factor in a second compute pass
// over the shaded buffer (render/gpudeferred.go shadowKernel).
func TestGPUDeferredShadow(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 200, 200
	s := scene.NewScene(
		light.NewPoint(light.Intensity(3), light.Position(math.NewVec3[float32](4, 4, 2)), light.CastShadow(true)),
		light.NewAmbient(light.Intensity(0.7)),
	)
	m := model.MustLoad("../internal/testdata/bunny.obj")
	m.Scale(2, 2, 2)
	s.Add(m)
	g := model.MustLoad("../internal/testdata/ground.obj")
	g.Scale(2, 2, 2)
	s.Add(g)
	scene.IterObjects(s, func(o *geometry.Geometry, _ math.Mat4[float32]) bool {
		for _, mid := range o.Materials() {
			material.Get(mid).Config(material.ReceiveShadow(true))
		}
		return true
	})
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 0.6, 0.9)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 2),
	)

	opts := []Option{
		Camera(cam), Size(w, h), MSAA(1), Scene(s), ShadowMap(true),
		Background(color.RGBA{R: 0, G: 127, B: 255, A: 255}), Workers(1), BatchSize(1),
	}

	cpu := NewRenderer(opts...).Render()
	gpuDeferredUsed = false
	gpuImg := NewRenderer(append(opts, GPU(dev))...).Render()
	if !gpuDeferredUsed {
		t.Fatal("GPU deferred path not exercised (shadow)")
	}

	maxDiff, at := 0, 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > maxDiff {
			maxDiff, at = d, i
		}
	}
	if maxDiff > 2 {
		px, py := (at/4)%w, (at/4)/w
		t.Fatalf("CPU vs GPU deferred (shadow): max diff = %d at (%d,%d) cpu=%v gpu=%v",
			maxDiff, px, py, cpu.Pix[at/4*4:at/4*4+4], gpuImg.Pix[at/4*4:at/4*4+4])
	}
	t.Logf("GPU deferred shadow parity: max channel diff = %d", maxDiff)
}
