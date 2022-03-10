// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
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
)

func NewShadowScene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene()

	s.Add(
		light.NewPoint(
			light.Intensity(3),
			light.Position(math.NewVec3[float32](4, 4, 2)),
			light.CastShadow(true),
		),
		light.NewPoint(
			light.Intensity(3),
			light.Position(math.NewVec3[float32](-6, 4, 2)),
			light.CastShadow(true),
		),
		light.NewAmbient(
			light.Intensity(0.7),
		),
	)

	m, err := mesh.LoadAs[*mesh.TriangleMesh]("../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}

	data := imageutil.MustLoadImage("../testdata/bunny.png")
	tex := buffer.NewTexture(
		buffer.TextureImage(data),
		buffer.TextureIsoMipmap(true),
	)
	mat := material.NewBlinnPhong(
		material.Texture(tex),
		material.Kdiff(0.6), material.Kspec(0.3),
		material.Shininess(20),
	)
	m.SetMaterial(mat)
	m.Scale(2, 2, 2)
	s.Add(m)

	m, err = mesh.LoadAs[*mesh.TriangleMesh]("../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	data = imageutil.MustLoadImage("../testdata/ground.png")
	tex = buffer.NewTexture(
		buffer.TextureImage(data),
		buffer.TextureIsoMipmap(true),
	)
	mat = material.NewBlinnPhong(
		material.Texture(tex),
		material.Kdiff(0.6), material.Kspec(0.3),
		material.Shininess(20),
		material.ReceiveShadow(true),
	)
	m.SetMaterial(mat)
	m.Scale(2, 2, 2)
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 0.6, 0.9)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 2),
	)
}

func TestShadow(t *testing.T) {
	tests := []*BasicOpt{
		{
			Name:       "shadow",
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
		s, cam := NewShadowScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)
		Render(t, s, test, rendopts...)
	}
}
