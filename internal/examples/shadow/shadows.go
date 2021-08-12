// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import (
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene"
	"poly.red/texture"
)

func NewShadowScene(w, h int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.Position(math.NewVec3(0, 0.6, 0.9)),
		camera.PerspFrustum(45, float64(w)/float64(h), 0.1, 2),
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

	m, err := mesh.Load("../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}

	data := texture.MustLoadImage("../testdata/bunny.png")
	tex := texture.NewTexture(
		texture.WithSource(data),
		texture.WithIsotropicMipMap(true),
	)
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 0.3),
		material.WithBlinnPhongShininess(20),
	)
	m.SetMaterial(mat)
	m.Scale(2, 2, 2)
	s.Add(m)

	m, err = mesh.Load("../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	data = texture.MustLoadImage("../testdata/ground.png")
	tex = texture.NewTexture(
		texture.WithSource(data),
		texture.WithIsotropicMipMap(true),
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
