// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"
	"image/color"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

// See more mesh data here: https://casual-effects.com/data/

func main() {
	width, height, msaa := 960, 540, 2
	s := rend.NewScene()
	c := camera.NewPerspectiveCamera(
		math.NewVector(-3, 1.25, -2, 1),
		math.NewVector(0, -0.1, -0.1, 1),
		math.NewVector(0, 1, 0, 0),
		30, float64(width)/float64(height), 0.01, 1000,
	)
	s.UseCamera(c)

	l := light.NewPointLight(20, color.RGBA{255, 255, 255, 255}, math.NewVector(-1.5, -1, 1, 1))
	s.AddLight(l)

	var done func()

	// load a mesh
	done = utils.Timed("loading mesh")
	m := io.MustLoadMesh("../../testdata/dragon.obj")
	done()

	done = utils.Timed("loading texture")
	data := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data.Set(0, 0, color.RGBA{0, 128, 255, 255})
	tex := material.NewTexture(data, true)
	done()

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.5, 0.6, 200),
		material.WithBlinnPhongShininess(25),
	)
	m.UseMaterial(mat)
	m.Scale(1.5, 1.5, 1.5)
	m.Translate(0, -0.1, -0.15)
	s.AddMesh(m)

	r := rend.NewRenderer(
		rend.WithSize(width, height),
		rend.WithMSAA(msaa),
		rend.WithScene(s),
		rend.WithDebug(true),
	)

	done = utils.Timed("rendering")
	buf := r.Render()
	done()

	done = utils.Timed("save")
	utils.Save(buf, "dragon.png")
	done()
}
