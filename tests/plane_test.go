// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests

import (
	"testing"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/color"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

func TestPlane(t *testing.T) {
	s := rend.NewScene()
	c := camera.NewPerspectiveCamera(
		math.NewVector(2, 2, 2, 1),
		math.NewVector(0, 0, 0, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		1,
		0.1, 10,
	)
	s.UseCamera(c)

	l := light.NewPointLight(
		1,
		color.RGBA{0, 128, 255, 255},
		math.NewVector(2, 2, 2, 1),
	)
	s.AddLight(l)

	m := geometry.NewPlane()
	s.AddMesh(m)

	r := rend.NewRenderer(
		rend.WithSize(500, 500),
		rend.WithMSAA(2),
		rend.WithScene(s),
		rend.WithBackground(color.FromHex("#181818")),
	)
	utils.Save(r.Render(), "plane.png")
}
