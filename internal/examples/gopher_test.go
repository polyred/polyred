// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"image/color"
	"testing"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
)

func NewGopherScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.Intensity(5),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3[float32](0, 0, 5)),
	), light.NewAmbient(
		light.Intensity(0.7),
	))

	m, err := mesh.LoadAs[*mesh.TriangleMesh]("../testdata/gopher.obj")
	if err != nil {
		panic(err)
	}
	m.RotateY(-math.Pi / 2)

	mat := material.NewBlinnPhong(
		material.Texture(buffer.NewUniformTexture(color.RGBA{0, 128, 255, 255})),
		material.Kdiff(0.6), material.Kspec(1),
		material.Shininess(150),
		material.FlatShading(true),
		material.AmbientOcclusion(true),
	)
	m.SetMaterial(mat)
	m.Normalize()
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
