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

// TestGPUDeferredAO offloads screen-space ambient occlusion to the GPU. The
// engine's SSAO ends in pow(total, 10000), which amplifies any GPU/CPU float
// difference (atan/cos/sin), so exact parity is not expected; this asserts the
// images are *close* (and reports the actual max diff).
func TestGPUDeferredAO(t *testing.T) {
	dev, err := gpu.Open()
	if err != nil {
		t.Skipf("no GPU device: %v", err)
	}
	defer dev.Close()

	const w, h = 150, 150
	s := scene.NewScene(
		light.NewPoint(light.Intensity(3), light.Position(math.NewVec3[float32](2, 3, 4))),
		light.NewAmbient(light.Intensity(0.5)),
	)
	m := model.MustLoad("../internal/testdata/bunny.obj")
	m.Scale(2, 2, 2)
	s.Add(m)
	scene.IterObjects(s, func(o *geometry.Geometry, _ math.Mat4[float32]) bool {
		for _, mid := range o.Materials() {
			material.Get(mid).Config(material.AmbientOcclusion(true))
		}
		return true
	})
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 0.6, 0.9)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 2),
	)
	opts := []Option{Camera(cam), Size(w, h), MSAA(1), Scene(s), Background(color.RGBA{R: 0, G: 127, B: 255, A: 255}), Workers(1), BatchSize(1)}

	cpu := NewRenderer(opts...).Render()
	gpuDeferredUsed = false
	gpuImg := NewRenderer(append(opts, GPU(dev))...).Render()
	if !gpuDeferredUsed {
		t.Fatal("GPU deferred path not exercised (AO)")
	}

	maxDiff, nBig := 0, 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > maxDiff {
			maxDiff = d
		}
		if d > 8 {
			nBig++
		}
	}
	t.Logf("GPU SSAO vs CPU: max channel diff = %d, channels differing by >8 = %d/%d", maxDiff, nBig, len(cpu.Pix))
	// SSAO's pow(.,10000) makes a handful of edge pixels diverge; require the
	// vast majority to be close.
	if frac := float64(nBig) / float64(len(cpu.Pix)); frac > 0.02 {
		t.Fatalf("too many large SSAO differences: %.2f%% of channels differ by >8", frac*100)
	}
}
