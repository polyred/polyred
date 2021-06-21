// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package dragon

import (
	"image/color"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/image"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/scene"
)

func NewDragonScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVector(-3, 1.25, -2, 1),
		math.NewVector(0, -0.1, -0.1, 1),
		math.NewVector(0, 1, 0, 0),
		30, float64(w)/float64(h), 0.01, 1000,
	))

	s.Add(light.NewPoint(
		light.WithPointLightIntensity(5),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVector(-1.5, -1, 1, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.2),
	))

	m := io.MustLoadMesh("../testdata/dragon.obj")
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			image.NewColorTexture(color.RGBA{0, 128, 255, 255}),
		),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(100),
	))
	m.Scale(1.5, 1.5, 1.5)
	m.Translate(0, -0.1, -0.15)
	s.Add(m)

	return s
}
