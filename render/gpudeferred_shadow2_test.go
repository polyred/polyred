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

// TestGPUDeferredShadowTwoLights exercises the multi-light shadow path: two
// shadow-casting lights (as in the engine's shadow example). The kernel loops
// over both per-light matrices/depth maps.
func TestGPUDeferredShadowTwoLights(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 200, 200
	s := scene.NewScene(
		light.NewPoint(light.Intensity(3), light.Position(math.NewVec3[float32](4, 4, 2)), light.CastShadow(true)),
		light.NewPoint(light.Intensity(3), light.Position(math.NewVec3[float32](-6, 4, 2)), light.CastShadow(true)),
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
	gr := NewRenderer(append(opts, GPU(dev))...)
	gpuImg := gr.Render()
	if !gr.passOnGPU("deferred") {
		t.Fatal("GPU deferred path not exercised (two-light shadow)")
	}

	assertDeferredClose(t, cpu.Pix, gpuImg.Pix, "two-light shadow")
}
