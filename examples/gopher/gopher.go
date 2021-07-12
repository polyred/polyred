// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gopher

import (
	"image/color"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/scene"
)

func NewGopherScene(width, height int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVec3(1, 1, 2),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 1, 0),
		45,
		float64(width)/float64(height),
		0.01, 600,
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(5),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(0, 0, 5)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	m := io.MustLoadMesh("../testdata/gopher.obj")
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
