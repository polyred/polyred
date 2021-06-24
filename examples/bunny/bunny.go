// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package bunny

import (
	"image/color"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
)

func NewBunnyScene(width, height int) interface{} {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVector(-550, 194, 734, 1),
		math.NewVector(-1000, 0, 0, 1),
		math.NewVector(0, 1, 1, 0),
		45,
		float64(width)/float64(height),
		100, 600,
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(200),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVector(-200, 250, 600, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	var done func()

	// load a mesh
	done = utils.Timed("loading mesh")
	m := io.MustLoadMesh("../testdata/bunny-smooth.obj")
	done()

	done = utils.Timed("loading texture")
	data := io.MustLoadImage("../testdata/bunny.png")
	tex := image.NewTexture(
		image.WithSource(data),
		image.WithIsotropicMipMap(true),
	)
	done()

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
		material.WithBlinnPhongFlatShading(true),
	)
	m.SetMaterial(mat)
	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.Add(m)

	return s
}
