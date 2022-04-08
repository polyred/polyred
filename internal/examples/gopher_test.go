// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func NewGopherScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(light.NewPoint(
		light.Intensity(5),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3[float32](0, 0, 5)),
	), light.NewAmbient(
		light.Intensity(0.7),
	))

	m := model.MustLoadAs[*mesh.TriangleMesh]("../testdata/gopher.obj")
	m.RotateY(-math.Pi / 2)
	// m.Normalize()
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](1, 1, 2)),
		camera.ViewFrustum(45, float32(width)/float32(height), 0.01, 600),
	)
}

func TestGopher(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "gopher",
			Width:      500,
			Height:     500,
			CPUProf:    false,
			MemProf:    false,
			ExecTracer: false,
			RenderOpts: []render.Option{
				render.Debug(false),
				render.MSAA(1),
				render.ShadowMap(true),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewGopherScene(test.Width, test.Height)
		rendopts := []render.Option{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
