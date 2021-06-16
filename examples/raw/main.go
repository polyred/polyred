// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"log"
	"os"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

func main() {
	width, height, msaa := 1000, 1000, 4
	models := []string{
		"plane",
		"cube",
		"cone",
		"cylinder",
		"ico",
		"torus",
		"knot",
		"sphere",
		"bunny",
		"monkey",
		"dragon",
		"teapot",
		"buddha",
		"conference",
		"roadBike",
		"hairball",
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot get home dir: %v", err)
	}

	for _, model := range models {
		s := rend.NewScene()
		c := camera.NewPerspective(
			math.NewVector(1, 1, 2, 1),
			math.NewVector(0, 0, 0, 1),
			math.NewVector(0, 1, 0, 0),
			50, 1, 0.1, 100,
		)
		s.UseCamera(c)
		s.AddLight(light.NewPointLight(
			10,
			color.RGBA{255, 255, 255, 255},
			math.NewVector(2, 2, 2, 1),
		))

		m := io.MustLoadMesh(fmt.Sprintf("%s/Dropbox/Data/%s.obj", home, model))
		m.Normalize()
		m.UseMaterial(material.NewBasicMaterial(color.RGBA{0, 128, 255, 255}))
		s.AddMesh(m)

		r := rend.NewRenderer(
			rend.WithSize(width, height),
			rend.WithMSAA(msaa),
			rend.WithScene(s),
		)

		fmt.Printf("rendering: %s\n", model)
		buf := r.Render()
		utils.Save(buf, fmt.Sprintf("%s/Dropbox/Data/%s.png", home, model))
	}
}
