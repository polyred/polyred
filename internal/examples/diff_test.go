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
	"poly.red/texture/imageutil"
)

func NewDiffScene(width, height int, lightI float32) (*scene.Scene, camera.Interface) {
	// Create a scene graph
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.Intensity(lightI),
		light.Color(color.RGBA{255, 255, 255, 255}),
		light.Position(math.NewVec3(4, 4, 2)),
		light.CastShadow(true)),
		light.NewAmbient(light.Intensity(0.5)))

	m, err := mesh.Load("../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.Texture(
			texture.NewTexture(
				texture.Image(
					imageutil.MustLoadImage("../testdata/bunny.png",
						imageutil.GammaCorrect(true)),
				),
				texture.IsoMipmap(true),
			),
		),
		material.Kdiff(0.6), material.Kspec(1),
		material.Shininess(150),
		material.ReceiveShadow(true),
		material.AmbientOcclusion(true),
	))
	m.Scale(2, 2, 2)
	s.Add(m)

	m, err = mesh.Load("../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	m.SetMaterial(material.NewBlinnPhong(
		material.Texture(texture.NewTexture(
			texture.Image(
				imageutil.MustLoadImage("../testdata/ground.png",
					imageutil.GammaCorrect(true)),
			),
			texture.IsoMipmap(true),
		),
		),
		material.Kdiff(0.6), material.Kspec(0.5),
		material.Shininess(150),
		material.ReceiveShadow(true),
	))
	m.Scale(2, 2, 2)
	s.Add(m)

	return s, camera.NewPerspective(
		camera.Position(math.NewVec3(0, 0.6, 0.9)),
		camera.ViewFrustum(45, float32(width)/float32(height), 0.1, 2),
	)
}

func TestDiff(t *testing.T) {
	// We don't run this test for now.
	t.Skip()

	width, height, msaa := 500, 500, 1
	Igoal := 7
	goal, cam := NewDiffScene(width, height, float32(Igoal))
	goalR := render.NewRenderer(
		render.Camera(cam),
		render.Size(width, height),
		render.MSAA(msaa),
		render.Scene(goal),
		render.ShadowMap(true),
		render.GammaCorrection(true),
	)
	goalImg := goalR.Render()
	imageutil.Save(goalImg, "./out/goal.png")

	searchR := render.NewRenderer(
		render.Size(width, height),
		render.MSAA(msaa),
		render.ShadowMap(true),
		render.GammaCorrection(true),
	)

	// naive random search.
	iter := 0
	for {
		Isearch := rand.Float32() * 10
		searchS, cam := NewDiffScene(width, height, Isearch)
		searchR.Options(render.Camera(cam), render.Scene(searchS))
		searchImg := searchR.Render()
		imageutil.Save(searchImg, "./out/search.png")
		diffImg, diffScore := imageutil.Diff(goalImg, searchImg, imageutil.MseKernel)
		imageutil.Save(diffImg, fmt.Sprintf("./out/diff-%d-search-%f-score-%f.png", iter, Isearch, diffScore))
		iter++
		if diffScore < 1000 {
			break
		}
	}
}
