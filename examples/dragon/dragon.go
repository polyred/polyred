// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package dragon

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

func NewDragonScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVec4(-3, 1.25, -2, 1),
		math.NewVec4(0, -0.1, -0.1, 1),
		math.NewVec4(0, 1, 0, 0),
		30, float64(w)/float64(h), 0.01, 1000,
	))

	s.Add(light.NewPoint(
		light.WithPointLightIntensity(2),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec4(-1.5, -1, 1, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.5),
	))

	m := io.MustLoadMesh("../testdata/dragon.obj")
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			image.NewColorTexture(color.RGBA{0, 128, 255, 255}),
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
