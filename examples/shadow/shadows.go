// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import (
	"poly.red/camera"
	"poly.red/image"
	"poly.red/io"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
)

func NewShadowScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.WithPosition(math.NewVec3(0, 0.6, 0.9)),
		camera.WithPerspFrustum(45, float64(w)/float64(h), 0.1, 2),
	))

	s.Add(
		light.NewPoint(
			light.WithPointLightIntensity(3),
			light.WithPointLightPosition(math.NewVec3(4, 4, 2)),
			light.WithPointLightShadowMap(true),
		),
		light.NewPoint(
			light.WithPointLightIntensity(3),
			light.WithPointLightPosition(math.NewVec3(-6, 4, 2)),
			light.WithPointLightShadowMap(true),
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
