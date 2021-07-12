// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests

import (
	"testing"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
)

func TestPlane(t *testing.T) {
	s := scene.NewScene()
	s.SetCamera(camera.NewPerspective(
		math.NewVec3(2, 2, 2),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 1, 0),
		45,
		1,
		0.1, 10,
	))
	s.Add(light.NewPoint(
		light.WithPointLightIntensity(1),
		light.WithPointLightColor(color.RGBA{0, 128, 255, 255}),
		light.WithPointLightPosition(math.NewVec3(2, 2, 2)),
	))

	m := geometry.NewPlane(1, 1)
	s.Add(m)

	r := render.NewRenderer(
		render.WithSize(500, 500),
		render.WithMSAA(2),
		render.WithScene(s),
		render.WithBackground(color.FromHex("#181818")),
	)
	utils.Save(r.Render(), "plane.png")
}
