// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"
)

func NewDragonScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.Intensity(2),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3(-1.5, -1, 1)),
	), light.NewAmbient(
		light.Intensity(0.5),
	))

	m, err := mesh.Load("../testdata/dragon.obj")
	if err != nil {
		panic(err)
	}

	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			texture.NewColorTexture(color.RGBA{0, 128, 255, 255}),
		),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(100),
		material.WithBlinnPhongAmbientOcclusion(true),
	))
	m.Scale(1.5, 1.5, 1.5)
	m.Translate(0, -0.1, -0.15)
	m.Normalize()
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3(-3, 1.25, -2)),
		camera.LookAt(math.NewVec3(0, -0.1, -0.1), math.NewVec3(0, 1, 0)),
		camera.ViewFrustum(30, float64(w)/float64(h), 0.01, 1000),
	)
}

func TestDragonScene(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "dragon",
			Width:      500,
			Height:     500,
			CPUProf:    false,
			MemProf:    false,
			ExecTracer: false,
			RenderOpts: []render.Opt{
				render.MSAA(2),
				render.ShadowMap(false),
				render.Debug(true),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewDragonScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
