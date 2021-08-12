// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"math/rand"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"

	"poly.red/internal/utils"
)

func loadScene(width, height int, lightI float64) *scene.Scene {
	// Create a scene graph
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		camera.Position(math.NewVec3(0, 0.6, 0.9)),
		camera.PerspFrustum(45, float64(width)/float64(height), 0.1, 2),
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(lightI),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(4, 4, 2)),
		light.WithPointLightShadowMap(true)),
		light.NewAmbient(light.WithAmbientIntensity(0.5)))

	m, err := mesh.Load("../../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			texture.NewTexture(
				texture.WithSource(
					texture.MustLoadImage("../../testdata/bunny.png",
						texture.WithGammaCorrection(true)),
				),
				texture.WithIsotropicMipMap(true),
			),
		),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongShadow(true),
		material.WithBlinnPhongAmbientOcclusion(true),
	))
	m.Scale(2, 2, 2)
	s.Add(m)

	m, err = mesh.Load("../../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(texture.NewTexture(
			texture.WithSource(
				texture.MustLoadImage("../../testdata/ground.png",
					texture.WithGammaCorrection(true)),
			),
			texture.WithIsotropicMipMap(true),
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
		diffImg, diffScore := texture.MseDiff(goalImg, searchImg)
		utils.Save(diffImg, fmt.Sprintf("diff-%d-search-%f-score-%f.png", iter, Isearch, diffScore))
		iter++
		if diffScore < 1000 {
			break
		}
	}
}
