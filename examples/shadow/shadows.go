// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image/color"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

func main() {
	width, height, msaa, shadow := 960, 540, 3, true
	s := rend.NewScene()

	c := camera.NewPerspective(
		math.NewVector(0, 0.6, 0.9, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		0.1,
		2,
	)
	s.UseCamera(c)

	s.AddLight(light.NewPoint(
		light.WithPoingLightItensity(20),
		light.WithPoingLightColor(color.RGBA{0, 0, 0, 255}),
		light.WithPoingLightPosition(math.NewVector(4, 4, 2, 1)),
		light.WithShadowMap(true),
	))
	s.AddLight(light.NewPoint(
		light.WithPoingLightItensity(20),
		light.WithPoingLightColor(color.RGBA{0, 0, 0, 255}),
		light.WithPoingLightPosition(math.NewVector(-4, 4, -2, 1)),
		light.WithShadowMap(true),
	))

	m := io.MustLoadMesh("../../testdata/bunny.obj")
	data := io.MustLoadImage("../../testdata/bunny.png")
	tex := material.NewTexture(
		material.WithImage(data),
		material.WithIsotropicMipMap(true),
	)
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.5, 0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.UseMaterial(mat)
	m.Scale(2, 2, 2)
	s.AddMesh(m)

	m = io.MustLoadMesh("../../testdata/ground.obj")
	data = io.MustLoadImage("../../testdata/ground.png")
	tex = material.NewTexture(
		material.WithImage(data),
		material.WithIsotropicMipMap(true),
	)
	mat = material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.5, 0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongShadow(true),
	)
	m.UseMaterial(mat)
	m.Scale(2, 2, 2)
	s.AddMesh(m)

	r := rend.NewRenderer(
		rend.WithSize(width, height),
		rend.WithMSAA(msaa),
		rend.WithScene(s),
		rend.WithShadowMap(shadow),
		rend.WithDebug(true),
	)
	utils.Save(r.Render(), "./shadows.png")
}
