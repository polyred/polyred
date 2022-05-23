// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"

	"poly.red/internal/profiling"
)

func NewBunnyScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(light.NewPoint(
		light.Intensity(200),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3[float32](-200, 250, 600)),
	), light.NewAmbient(
		light.Intensity(0.7),
	))

	done := profiling.Timed("loading obj")
	m := model.MustLoad("../testdata/bunny.obj")
	done()

	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](-550, 194, 734)),
		camera.LookAt(math.NewVec3[float32](-1000, 0, 0), math.NewVec3[float32](0, 1, 1)),
		camera.ViewFrustum(45, float32(width)/float32(height), 100, 600),
	)
}

func TestBunny(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "bunny",
			Width:      960,
			Height:     540,
			CPUProf:    false,
			MemProf:    false,
			ExecTracer: false,
			RenderOpts: []render.Option{
				render.Debug(true),
				render.MSAA(2),
				render.ShadowMap(true),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewBunnyScene(test.Width, test.Height)
		rendopts := []render.Option{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
