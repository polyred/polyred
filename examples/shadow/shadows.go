// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import (
	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/scene"
)

func NewShadowScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVector(0, 0.6, 0.9, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(w)/float64(h),
		0.1,
		2,
	))

	s.Add(
		light.NewPoint(
			light.WithPointLightIntensity(3),
			light.WithPointLightPosition(math.NewVector(4, 4, 2, 1)),
			light.WithShadowMap(true),
		),
		light.NewPoint(
			light.WithPointLightIntensity(3),
			light.WithPointLightPosition(math.NewVector(-6, 4, 2, 1)),
			light.WithShadowMap(true),
		),
		light.NewAmbient(
			light.WithAmbientIntensity(0.7),
		),
	)

	m := io.MustLoadMesh("../testdata/bunny.obj")
	data := io.MustLoadImage("../testdata/bunny.png")
	tex := image.NewTexture(
		image.WithSource(data),
		image.WithIsotropicMipMap(true),
	)
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 0.3),
		material.WithBlinnPhongShininess(20),
	)
	m.SetMaterial(mat)
	m.Scale(2, 2, 2)
	s.Add(m)

	m = io.MustLoadMesh("../testdata/ground.obj")
	data = io.MustLoadImage("../testdata/ground.png")
	tex = image.NewTexture(
		image.WithSource(data),
		image.WithIsotropicMipMap(true),
	)
	mat = material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 0.3),
		material.WithBlinnPhongShininess(20),
		material.WithBlinnPhongShadow(true),
	)
	m.SetMaterial(mat)
	m.Scale(2, 2, 2)
	s.Add(m)

	return s
}
