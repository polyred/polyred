// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"math/rand"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
)

func loadScene(width, height int, lightI float64) *scene.Scene {
	// Create a scene graph
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVector(0, 0.6, 0.9, 1), // position
		math.NewVector(0, 0, 0, 1),     // lookAt
		math.NewVector(0, 1, 0, 0),     // up
		45,                             // fov
		float64(width)/float64(height), // aspect
		0.1,                            // near
		2,                              // far
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(lightI),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVector(4, 4, 2, 1)),
		light.WithPointLightShadowMap(true)),
		light.NewAmbient(light.WithAmbientIntensity(0.5)))

	m := io.MustLoadMesh("../../testdata/bunny.obj")
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			image.NewTexture(
				image.WithSource(
					io.MustLoadImage("../../testdata/bunny.png",
						io.WithGammaCorrection(true)),
				),
				image.WithIsotropicMipMap(true),
			),
		),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongShadow(true),
		material.WithBlinnPhongAmbientOcclusion(true),
	))
	m.Scale(2, 2, 2)
	s.Add(m)

	m = io.MustLoadMesh("../../testdata/ground.obj")
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(image.NewTexture(
			image.WithSource(
				io.MustLoadImage("../../testdata/ground.png",
					io.WithGammaCorrection(true)),
			),
			image.WithIsotropicMipMap(true),
		),
		),
		material.WithBlinnPhongFactors(0.6, 0.5),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongShadow(true),
	))
	m.Scale(2, 2, 2)
	s.Add(m)

	return s
}

func main() {
	width, height, msaa := 500, 500, 1
	Igoal := 7
	goal := loadScene(width, height, float64(Igoal))
	goalR := render.NewRenderer(
		render.WithSize(width, height),
		render.WithMSAA(msaa),
		render.WithScene(goal),
		render.WithShadowMap(true),
		render.WithGammaCorrection(true),
	)
	goalImg := goalR.Render()
	utils.Save(goalImg, "goal.png")

	searchR := render.NewRenderer(
		render.WithSize(width, height),
		render.WithMSAA(msaa),
		render.WithShadowMap(true),
		render.WithGammaCorrection(true),
	)

	// naive random search.
	iter := 0
	for {
		Isearch := rand.Float64() * 10
		searchS := loadScene(width, height, Isearch)
		searchR.UpdateOptions(
			render.WithScene(searchS),
		)
		searchImg := searchR.Render()
		utils.Save(searchImg, "search.png")
		diffImg, diffScore := image.MseDiff(goalImg, searchImg)
		utils.Save(diffImg, fmt.Sprintf("diff-%d-search-%f-score-%f.png", iter, Isearch, diffScore))
		iter++
		if diffScore < 1000 {
			break
		}
	}
}
