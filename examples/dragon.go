// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

// See more mesh data here: https://casual-effects.com/data/

func timed(name string) func() {
	if len(name) > 0 {
		fmt.Printf("%s... ", name)
	}
	start := time.Now()
	return func() {
		fmt.Println(time.Since(start))
	}
}

func main() {
	width, height, msaa := 2048, 2048, 2
	s := rend.NewScene()
	c := camera.NewPerspectiveCamera(
		math.NewVector(-3, 1.25, -2, 1),
		math.NewVector(0, -0.1, -0.1, 1),
		math.NewVector(0, 1, 0, 0),
		30, float64(width)/float64(height), -1, -10,
	)
	s.UseCamera(c)

	l := light.NewPointLight(color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
	s.AddLight(l)

	var done func()

	// load a mesh
	done = timed("loading mesh")
	m := geometry.MustLoad("./dragon.obj")
	done()

	done = timed("loading texture")

	data := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data.Set(0, 0, color.RGBA{0, 128, 255, 255})
	tex := material.NewTexture(data)
	done()

	mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 1, 1, 0.5, 150)
	m.UseMaterial(mat)
	m.Scale(1.5, 1.5, 1.5)
	s.AddMesh(m)

	r := rend.NewRasterizer(width, height, msaa)

	done = timed("rendering")
	buf := r.Render(s)
	done()

	utils.Save(buf, "dragon.png")
}
