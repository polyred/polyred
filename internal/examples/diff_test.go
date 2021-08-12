// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"fmt"
	"image/color"
	"math/rand"
	"testing"

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

func NewDiffScene(width, height int, lightI float64) (*scene.Scene, camera.Interface) {
	// Create a scene graph
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(lightI),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(4, 4, 2)),
		light.WithPointLightShadowMap(true)),
		light.NewAmbient(light.WithAmbientIntensity(0.5)))

	m, err := mesh.Load("../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(
			texture.NewTexture(
				texture.WithSource(
					texture.MustLoadImage("../testdata/bunny.png",
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

	m, err = mesh.Load("../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.WithBlinnPhongTexture(texture.NewTexture(
			texture.WithSource(
				texture.MustLoadImage("../testdata/ground.png",
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

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3(0, 0.6, 0.9)),
		camera.ViewFrustum(45, float64(width)/float64(height), 0.1, 2),
	)
}

func TestDiff(t *testing.T) {
	// We don't run this test for now.
	t.Skip()

	width, height, msaa := 500, 500, 1
	Igoal := 7
	goal, cam := NewDiffScene(width, height, float64(Igoal))
	goalR := render.NewRenderer(
		render.Camera(cam),
		render.Size(width, height),
		render.MSAA(msaa),
		render.Scene(goal),
		render.ShadowMap(true),
		render.GammaCorrection(true),
	)
	goalImg := goalR.Render()
	utils.Save(goalImg, "./out/goal.png")

	searchR := render.NewRenderer(
		render.Size(width, height),
		render.MSAA(msaa),
		render.ShadowMap(true),
		render.GammaCorrection(true),
	)

	// naive random search.
	iter := 0
	for {
		Isearch := rand.Float64() * 10
		searchS, cam := NewDiffScene(width, height, Isearch)
		searchR.Options(render.Camera(cam), render.Scene(searchS))
		searchImg := searchR.Render()
		utils.Save(searchImg, "./out/search.png")
		diffImg, diffScore := texture.MseDiff(goalImg, searchImg)
		utils.Save(diffImg, fmt.Sprintf("./out/diff-%d-search-%f-score-%f.png", iter, Isearch, diffScore))
		iter++
		if diffScore < 1000 {
			break
		}
	}
}
