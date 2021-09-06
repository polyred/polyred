// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"fmt"
	"image/color"
	"log"
	"os"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"
)

type Scene struct {
	Name  string
	Scene *scene.Scene
}

func NewMcGuireScene(w, h int) ([]*Scene, camera.Interface) {
	models := []string{
		// "AL05a",
		// "AL05m",
		// "AL05y",
		// "altostratus00",
		// "altostratus01",
		// "bmw",
		// "breakfast_room",
		// "buddha",
		"bunny",
		// "cone",
		// "conference",
		// "CornellBox-Empty-CO",
		// "CornellBox-Empty-RG",
		// "CornellBox-Empty-Squashed",
		// "CornellBox-Empty-White",
		// "CornellBox-Glossy-Floor",
		// "CornellBox-Glossy",
		// "CornellBox-Mirror",
		// "CornellBox-Original",
		// "CornellBox-Sphere",
		// "CornellBox-Water",
		// "cube",
		// "cumulus00",
		// "cumulus01",
		// "cumulus02",
		"cylinder",
		"dragon",
		// "erato-1",
		// "fireplace_room",
		// "gallery",
		// "geodesic_classI_10",
		// "geodesic_classI_20",
		// "geodesic_classI_2",
		// "geodesic_classI_3",
		// "geodesic_classI_4",
		// "geodesic_classI_5",
		// "geodesic_classI_7",
		// "geodesic_classII_10_10",
		// "geodesic_classII_1_1",
		// "geodesic_classII_20_20",
		// "geodesic_classII_2_2",
		// "geodesic_classII_3_3",
		// "geodesic_classII_4_4",
		// "geodesic_classII_5_5",
		// "geodesic_classII_7_7",
		// "geodesic_classII_dual_1_1",
		// "geodesic_classII_dual_5_5",
		// "geodesic_classIII_10_1",
		// "geodesic_classIII_10_2",
		// "geodesic_classIII_10_3",
		// "geodesic_classIII_10_4",
		// "geodesic_classIII_10_5",
		// "geodesic_classIII_10_7",
		// "geodesic_classIII_20_10",
		// "geodesic_classIII_20_1",
		// "geodesic_classIII_20_2",
		// "geodesic_classIII_20_3",
		// "geodesic_classIII_20_4",
		// "geodesic_classIII_20_5",
		// "geodesic_classIII_20_7",
		// "geodesic_classIII_2_1",
		// "geodesic_classIII_3_1",
		// "geodesic_classIII_3_2",
		// "geodesic_classIII_4_1",
		// "geodesic_classIII_4_2",
		// "geodesic_classIII_4_3",
		// "geodesic_classIII_5_1",
		// "geodesic_classIII_5_2",
		// "geodesic_classIII_5_3",
		// "geodesic_classIII_5_4",
		// "geodesic_classIII_7_1",
		// "geodesic_classIII_7_2",
		// "geodesic_classIII_7_3",
		// "geodesic_classIII_7_4",
		// "geodesic_classIII_7_5",
		// "geodesic_dual_classI_10",
		// "geodesic_dual_classI_20",
		// "geodesic_dual_classI_2",
		// "geodesic_dual_classI_3",
		// "geodesic_dual_classI_4",
		// "geodesic_dual_classI_5",
		// "geodesic_dual_classI_7",
		// "geodesic_dual_classII_10_10", // buggy?
		// "geodesic_dual_classII_20_20", // buggy?
		// "geodesic_dual_classII_2_2",
		// "geodesic_dual_classII_3_3",
		// "geodesic_dual_classII_4_4",
		// "geodesic_dual_classII_5_5",
		// "geodesic_dual_classII_7_7",
		// "geodesic_dual_classIII_10_1",
		// "geodesic_dual_classIII_10_2",
		// "geodesic_dual_classIII_10_3",
		// "geodesic_dual_classIII_10_4",
		// "geodesic_dual_classIII_10_5",
		// "geodesic_dual_classIII_10_7",
		// "geodesic_dual_classIII_20_10",
		// "geodesic_dual_classIII_20_1",
		// "geodesic_dual_classIII_20_2",
		// "geodesic_dual_classIII_20_3",
		// "geodesic_dual_classIII_20_4",
		// "geodesic_dual_classIII_20_5",
		// "geodesic_dual_classIII_20_7",
		// "geodesic_dual_classIII_2_1",
		// "geodesic_dual_classIII_3_1",
		// "geodesic_dual_classIII_3_2",
		// "geodesic_dual_classIII_4_1",
		// "geodesic_dual_classIII_4_2",
		// "geodesic_dual_classIII_4_3",
		// "geodesic_dual_classIII_5_1",
		// "geodesic_dual_classIII_5_2",
		// "geodesic_dual_classIII_5_3",
		// "geodesic_dual_classIII_5_4",
		// "geodesic_dual_classIII_7_1",
		// "geodesic_dual_classIII_7_3",
		// "geodesic_dual_classIII_7_4",
		// "geodesic_dual_classIII_7_5",
		// "hairball",
		// "holodeck",
		// "house",
		// "ico",
		// "iscv2",
		// "knot",
		// "living_room",
		// "lost_empire",
		// "mitsuba",
		// "mitsuba-sphere",
		"monkey",
		// "plane",
		// "powerplant",
		// "roadBike",
		// "rungholt",
		// "salle_de_bain",
		// "scrubPine",
		// "sibenik",
		// "sphere-cubecoords",
		// "sphere-cylcoords-16k",
		// "sphere-cylcoords-1k",
		// "sphere-cylcoords-4k",
		"sphere",
		// "sponza",
		// "sportsCar",
		"teapot",
		// "testObj",
		"torus",
		// "vokselia_spawn",
		// "water",
		// "white_oak",
	}
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot get home dir: %v", err)
	}

	scenes := make([]*Scene, len(models))
	for i, model := range models {
		s := &Scene{
			Scene: scene.NewScene(),
			Name:  model,
		}
		s.Scene.Add(light.NewPoint(
			light.Intensity(5),
			light.Color(color.RGBA{255, 255, 255, 255}),
			light.Position(math.NewVec3(2, 2, 2)),
		), light.NewAmbient(
			light.Intensity(0.5),
		))

		m, err := mesh.Load(fmt.Sprintf("%s/Dropbox/Data/%s.obj", home, model))
		if err != nil {
			panic(err)
		}

		m.Normalize()
		m.SetMaterial(material.NewBlinnPhong(
			material.Texture(texture.NewColorTexture(color.RGBA{0, 128, 255, 255})),
			material.Kdiff(0.6), material.Kspec(1),
			material.Shininess(100),
			material.FlatShading(false),
		))
		s.Scene.Add(m)

		scenes[i] = s
	}

	return scenes, camera.NewPerspective(
		camera.Position(math.NewVec3(1, 1, 2)),
		camera.ViewFrustum(50, float32(w)/float32(h), 0.1, 100),
	)
}

func TestMcguire(t *testing.T) {
	// FIXME: enable this test if we figured how to fetch data remotely.
	t.Skip()

	tests := []*BasicOpt{
		{
			Name:       "mcguire",
			Width:      540,
			Height:     540,
			CPUProf:    false,
			MemProf:    false,
			ExecTracer: false,
			RenderOpts: []render.Opt{
				render.Debug(true),
				render.MSAA(2),
				render.ShadowMap(false),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		ss, cam := NewMcGuireScene(test.Width, test.Height)
		rendopts := []render.Opt{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)

		rootName := test.Name
		for _, s := range ss {
			test.Name = rootName + "-" + s.Name
			Render(t, s.Scene, test, rendopts...)
		}
	}
}
