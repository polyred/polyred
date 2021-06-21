// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gopher

import (
	"image/color"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/image"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/scene"
	"changkun.de/x/ddd/utils"
)

func NewGopherScene(width, height int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVector(1, 1, 2, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		0.01, 600,
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(10),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVector(0, 0, 5, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	done := utils.Timed("loading mesh")
	m := io.MustLoadMesh("../testdata/gopher.obj")
	done()

	m.RotateY(-math.Pi / 2)

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(image.NewTexture()),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.SetMaterial(mat)
	m.Normalize()
	s.Add(m)

	return s
}
