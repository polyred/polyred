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
	width, height, msaa := 960, 540, 2
	s := rend.NewScene()
	c := camera.NewPerspective(
		math.NewVector(-550, 194, 734, 1),
		math.NewVector(-1000, 0, 0, 1),
		math.NewVector(0, 1, 1, 0),
		45,
		float64(width)/float64(height),
		100, 600,
	)
	s.UseCamera(c)
	s.AddLight(light.NewPoint(
		light.WithPointLightIntensity(200),
		light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
		light.WithPointLightPosition(math.NewVector(-200, 250, 600, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.7),
	))

	var done func()

	// load a mesh
	done = utils.Timed("loading mesh")
	m := io.MustLoadMesh("../../testdata/bunny.obj")
	done()

	done = utils.Timed("loading texture")
	data := io.MustLoadImage("../../testdata/bunny.png")
	tex := material.NewTexture(
		material.WithImage(data),
		material.WithIsotropicMipMap(true),
	)
	done()

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.UseMaterial(mat)
	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.AddMesh(m)

	r := rend.NewRenderer(
		rend.WithSize(width, height),
		rend.WithMSAA(msaa),
		rend.WithScene(s),
	)

	done = utils.Timed("rendering")
	buf := r.Render()
	done()

	utils.Save(buf, "bunny.png")
}
