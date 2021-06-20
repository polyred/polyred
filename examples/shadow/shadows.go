// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import (
	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
)

func NewShadowScene(w, h int) interface{} {
	s := rend.NewScene()
	s.UseCamera(camera.NewPerspective(
		math.NewVector(0, 0.6, 0.9, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(w)/float64(h),
		0.1,
		2,
	))

	s.AddLight(
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
	tex := material.NewTexture(
		material.WithImage(data),
		material.WithIsotropicMipMap(true),
	)
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 0.3),
		material.WithBlinnPhongShininess(20),
	)
	m.UseMaterial(mat)
	m.Scale(2, 2, 2)
	s.AddMesh(m)

	m = io.MustLoadMesh("../testdata/ground.obj")
	data = io.MustLoadImage("../testdata/ground.png")
	tex = material.NewTexture(
		material.WithImage(data),
		material.WithIsotropicMipMap(true),
	)
	mat = material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 0.3),
		material.WithBlinnPhongShininess(20),
		material.WithBlinnPhongShadow(true),
	)
	m.UseMaterial(mat)
	m.Scale(2, 2, 2)
	s.AddMesh(m)

	return s
}
