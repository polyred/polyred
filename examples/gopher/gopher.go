// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gopher

import (
	"image/color"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/image"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
)

func NewGopherScene(width, height int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.WithPosition(math.NewVec3(1, 1, 2)),
		camera.WithPerspFrustum(
			45,
			float64(width)/float64(height),
			0.01, 600,
		),
	))
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
		material.WithBlinnPhongTexture(image.NewColorTexture(color.RGBA{0, 128, 255, 255})),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongFlatShading(true),
		material.WithBlinnPhongAmbientOcclusion(true),
	)
	m.SetMaterial(mat)
	m.Normalize()
	s.Add(m)

	return s
}
