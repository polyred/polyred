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
	"poly.red/texture/imageutil"

	"poly.red/internal/profiling"
)

func NewBunnyScene(width, height int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.Intensity(200),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3(-200, 250, 600)),
	), light.NewAmbient(
		light.Intensity(0.7),
	))

	var done func()

	// load a mesh
	done = profiling.Timed("loading mesh")
	m, err := mesh.LoadAs[*mesh.TriangleSoup]("../testdata/bunny-smooth.obj")
	if err != nil {
		panic(err)
	}
	done()

	done = profiling.Timed("loading texture")
	data := imageutil.MustLoadImage("../testdata/bunny.png")
	tex := buffer.NewTexture(
		buffer.TextureImage(data),
		buffer.TextureIsoMipmap(true),
	)
	done()

	mat := material.NewBlinnPhong(
		material.Texture(tex),
		material.Kdiff(0.6), material.Kspec(1),
		material.Shininess(150),
		material.FlatShading(true),
	)
	m.SetMaterial(mat)
	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.Add(m)

	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(-550, 194, 734)),
		camera.LookAt(math.NewVec3(-1000, 0, 0), math.NewVec3(0, 1, 1)),
		camera.ViewFrustum(45, float32(width)/float32(height), 100, 600),
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
