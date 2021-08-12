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

func NewGopherScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(5),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(0, 0, 5)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	m, err := mesh.Load("../testdata/gopher.obj")
	if err != nil {
		panic(err)
	}
	m.RotateY(-math.Pi / 2)

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(texture.NewColorTexture(color.RGBA{0, 128, 255, 255})),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongFlatShading(true),
		material.WithBlinnPhongAmbientOcclusion(true),
	)
	m.SetMaterial(mat)
	m.Normalize()
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3(1, 1, 2)),
		camera.ViewFrustum(45, float64(width)/float64(height), 0.01, 600),
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
			RenderOpts: []render.Opt{
				render.Debug(false),
				render.MSAA(1),
				render.ShadowMap(true),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewGopherScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
