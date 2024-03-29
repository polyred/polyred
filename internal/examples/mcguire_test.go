// Copyright 2022 The Polyred Authors. All rights reserved.
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
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type Scene struct {
	Name  string
	Scene *scene.Scene
}

func NewMcGuireScene(t *testing.T, w, h int) ([]*Scene, camera.Interface) {
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
		"sponza",
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
	for i, mo := range models {
		s := &Scene{
			Scene: scene.NewScene(light.NewPoint(
				light.Intensity(5),
				light.Color(color.RGBA{255, 255, 255, 255}),
				light.Position(math.NewVec3[float32](2, 2, 2)),
			), light.NewAmbient(
				light.Intensity(0.5),
			)),
			Name: mo,
		}

		path := fmt.Sprintf("%s/Dropbox/Data/%s.obj", home, mo)
		m, err := model.Load(path)
		if err != nil {
			t.Skipf("cannot load model %s: %v", path, err)
		}
		m.Normalize()
		s.Scene.Add(m)

		scenes[i] = s
	}

	return scenes, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](1, 1, 2)),
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
			RenderOpts: []render.Option{
				render.Debug(true),
				render.MSAA(2),
				render.ShadowMap(false),
			},
		},
	}

	for _, test := range tests {
		t.Logf("%s under settings: %#v", test.Name, test)
		ss, cam := NewMcGuireScene(t, test.Width, test.Height)
		rendopts := []render.Option{render.Camera(cam)}
		rendopts = append(rendopts, test.RenderOpts...)

		rootName := test.Name
		for _, s := range ss {
			test.Name = rootName + "-" + s.Name
			Render(t, s.Scene, test, rendopts...)
		}
	}
}
