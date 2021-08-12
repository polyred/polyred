// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"image/color"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/internal/utils"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"
)

func NewBunnyScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(200),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(-200, 250, 600)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	var done func()

	// load a mesh
	done = utils.Timed("loading mesh")
	m, err := mesh.Load("../testdata/bunny-smooth.obj")
	if err != nil {
		panic(err)
	}
	done()

	done = utils.Timed("loading texture")
	data := texture.MustLoadImage("../testdata/bunny.png")
	tex := texture.NewTexture(
		texture.WithSource(data),
		texture.WithIsotropicMipMap(true),
	)
	done()

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongFlatShading(true),
	)
	m.SetMaterial(mat)
	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.Add(m)

	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(-550, 194, 734)),
		camera.LookAt(math.NewVec3(-1000, 0, 0), math.NewVec3(0, 1, 1)),
		camera.ViewFrustum(45, float64(width)/float64(height), 100, 600),
	)
	return s, cam
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
			RenderOpts: []render.Opt{
				render.Debug(false),
				render.MSAA(2),
				render.ShadowMap(true),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		s, cam := NewBunnyScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
