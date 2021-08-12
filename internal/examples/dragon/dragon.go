// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package dragon

import (
	"image/color"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
	"poly.red/texture"
)

func NewDragonScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.Position(math.NewVec3(-3, 1.25, -2)),
		camera.LookAt(math.NewVec3(0, -0.1, -0.1), math.NewVec3(0, 1, 0)),
		camera.PerspFrustum(30, float64(w)/float64(h), 0.01, 1000),
	))

	s.Add(light.NewPoint(
		light.WithPointLightIntensity(2),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(-1.5, -1, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.5),
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

	return s
}
