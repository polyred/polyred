// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests

import (
	"testing"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture/imageutil"
)

func TestPlane(t *testing.T) {
	s := scene.NewScene()
	s.Add(light.NewPoint(
		light.Intensity(1),
		light.Color(color.RGBA{0, 128, 255, 255}),
		light.Position(math.NewVec3(2, 2, 2)),
	))

	m := mesh.NewPlane(1, 1)
	s.Add(m)

	r := render.NewRenderer(
		render.Camera(camera.NewPerspective(
			camera.Position(math.NewVec3(2, 2, 2)),
			camera.ViewFrustum(45, 1, 0.1, 10),
		)),
		render.Size(500, 500),
		render.MSAA(2),
		render.Scene(s),
		render.Background(color.FromHex("#181818")),
	)
	imageutil.Save(r.Render(), "plane.png")
}
