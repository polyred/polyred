// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

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
	width, height, msaa := 800, 500, 1
	s := rend.NewScene()
	c := camera.NewPerspectiveCamera(
		math.Vector{-550, 194, 734, 1},
		math.Vector{-1000, 0, 0, 1},
		math.Vector{0, 1, 1, 0},
		45,
		float64(width)/float64(height),
		100, 600,
	)
	s.UseCamera(c)

	l := light.NewPointLight(
		200,
		color.RGBA{255, 255, 255, 255},
		math.Vector{-200, 250, 600, 1},
	)
	s.AddLight(l)

	var done func()

	// load a mesh
	done = utils.Timed("loading mesh")
	m := io.MustLoadMesh("../testdata/bunny.obj")
	done()

	done = utils.Timed("loading texture")
	t := io.MustLoadTexture("../testdata/bunny.png")
	done()

	mat := material.NewBlinnPhongMaterial(
		t,
		color.RGBA{0, 125, 255, 255},
		0.5,
		0.6,
		1,
		150,
	)
	m.UseMaterial(mat)
	m.Scale(1500, 1500, 1500)
	m.Translate(-700, -5, 350)
	s.AddMesh(m)

	r := rend.NewRasterizer(width, height, msaa)

	done = utils.Timed("rendering")
	buf := r.Render(s)
	done()

	utils.Save(buf, "bunny.png")
}
