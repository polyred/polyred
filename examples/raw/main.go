// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
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
	width, height, msaa := 500, 500, 1
	models := []string{"cube", "bunny", "monkey", "dragon"}

	for _, model := range models {
		s := rend.NewScene()
		c := camera.NewOrthographicCamera(
			math.NewVector(0, 0, 1, 1),
			math.NewVector(0, 0, -1, 1),
			math.NewVector(0, 1, 0, 0),
			-1, 1,
			-1, 1,
			1, -1,
		)
		s.UseCamera(c)
		s.AddLight(light.NewPointLight(
			10,
			color.RGBA{255, 255, 255, 255},
			math.NewVector(2, 2, 2, 1),
		))

		m := io.MustLoadMesh(fmt.Sprintf("./%s.obj", model))
		m.Normalize()
		m.UseMaterial(material.NewBasicMaterial(color.RGBA{0, 128, 255, 255}))
		s.AddMesh(m)

		r := rend.NewRenderer(
			rend.WithSize(width, height),
			rend.WithMSAA(msaa),
			rend.WithScene(s),
			rend.WithDebug(true),
		)

		fmt.Printf("rendering: %s\n", model)
		buf := r.Render()
		utils.Save(buf, fmt.Sprintf("./%s.png", model))
	}
}
