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
		"AL05a",
		"AL05m",
		"AL05y",
		"altostratus00",
		"altostratus01",
		"bmw",
		"breakfast_room",
		"buddha",
		"bunny",
		"cone",
		"conference",
		"CornellBox-Empty-CO",
		"CornellBox-Empty-RG",
		"CornellBox-Empty-Squashed",
		"CornellBox-Empty-White",
		"CornellBox-Glossy-Floor",
		"CornellBox-Glossy",
		"CornellBox-Mirror",
		"CornellBox-Original",
		"CornellBox-Sphere",
		"CornellBox-Water",
		"cube",
		"cumulus00",
		"cumulus01",
		"cumulus02",
		"cylinder",
		"dragon",
		"erato-1",
		"fireplace_room",
		"gallery",
		"geodesic_classI_10",
		"geodesic_classI_20",
		"geodesic_classI_2",
		"geodesic_classI_3",
		"geodesic_classI_4",
		"geodesic_classI_5",
		"geodesic_classI_7",
		"geodesic_classII_10_10",
		"geodesic_classII_1_1",
		"geodesic_classII_20_20",
		"geodesic_classII_2_2",
		"geodesic_classII_3_3",
		"geodesic_classII_4_4",
		"geodesic_classII_5_5",
		"geodesic_classII_7_7",
		"geodesic_classII_dual_1_1",
		"geodesic_classII_dual_5_5",
		"geodesic_classIII_10_1",
		"geodesic_classIII_10_2",
		"geodesic_classIII_10_3",
		"geodesic_classIII_10_4",
		"geodesic_classIII_10_5",
		"geodesic_classIII_10_7",
		"geodesic_classIII_20_10",
		"geodesic_classIII_20_1",
		"geodesic_classIII_20_2",
		"geodesic_classIII_20_3",
		"geodesic_classIII_20_4",
		"geodesic_classIII_20_5",
		"geodesic_classIII_20_7",
		"geodesic_classIII_2_1",
		"geodesic_classIII_3_1",
		"geodesic_classIII_3_2",
		"geodesic_classIII_4_1",
		"geodesic_classIII_4_2",
		"geodesic_classIII_4_3",
		"geodesic_classIII_5_1",
		"geodesic_classIII_5_2",
		"geodesic_classIII_5_3",
		"geodesic_classIII_5_4",
		"geodesic_classIII_7_1",
		"geodesic_classIII_7_2",
		"geodesic_classIII_7_3",
		"geodesic_classIII_7_4",
		"geodesic_classIII_7_5",
		"geodesic_dual_classI_10",
		"geodesic_dual_classI_20",
		"geodesic_dual_classI_2",
		"geodesic_dual_classI_3",
		"geodesic_dual_classI_4",
		"geodesic_dual_classI_5",
		"geodesic_dual_classI_7",
		"geodesic_dual_classII_10_10",
		// "geodesic_dual_classII_20_20",
		"geodesic_dual_classII_2_2",
		"geodesic_dual_classII_3_3",
		"geodesic_dual_classII_4_4",
		"geodesic_dual_classII_5_5",
		"geodesic_dual_classII_7_7",
		"geodesic_dual_classIII_10_1",
		"geodesic_dual_classIII_10_2",
		"geodesic_dual_classIII_10_3",
		"geodesic_dual_classIII_10_4",
		"geodesic_dual_classIII_10_5",
		"geodesic_dual_classIII_10_7",
		"geodesic_dual_classIII_20_10",
		"geodesic_dual_classIII_20_1",
		"geodesic_dual_classIII_20_2",
		"geodesic_dual_classIII_20_3",
		"geodesic_dual_classIII_20_4",
		"geodesic_dual_classIII_20_5",
		"geodesic_dual_classIII_20_7",
		"geodesic_dual_classIII_2_1",
		"geodesic_dual_classIII_3_1",
		"geodesic_dual_classIII_3_2",
		"geodesic_dual_classIII_4_1",
		"geodesic_dual_classIII_4_2",
		"geodesic_dual_classIII_4_3",
		"geodesic_dual_classIII_5_1",
		"geodesic_dual_classIII_5_2",
		"geodesic_dual_classIII_5_3",
		"geodesic_dual_classIII_5_4",
		"geodesic_dual_classIII_7_1",
		"geodesic_dual_classIII_7_3",
		"geodesic_dual_classIII_7_4",
		"geodesic_dual_classIII_7_5",
		"hairball",
		"holodeck",
		"house",
		"ico",
		"iscv2",
		"knot",
		"living_room",
		"lost_empire",
		"mitsuba",
		"mitsuba-sphere",
		"monkey",
		"plane",
		"powerplant",
		"roadBike",
		"rungholt",
		"salle_de_bain",
		"scrubPine",
		"sibenik",
		"sphere-cubecoords",
		"sphere-cylcoords-16k",
		"sphere-cylcoords-1k",
		"sphere-cylcoords-4k",
		"sphere",
		"sponza",
		"sportsCar",
		"teapot",
		"testObj",
		"torus",
		"vokselia_spawn",
		"water",
		"white_oak",
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
		s.AddLight(light.NewPoint(
			light.WithPointLightIntensity(5),
			light.WithPointLightColor(color.RGBA{255, 255, 255, 255}),
			light.WithPointLightPosition(math.NewVector(2, 2, 2, 1)),
		), light.NewAmbient(
			light.WithAmbientIntensity(0.5),
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
