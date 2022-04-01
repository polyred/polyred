// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/internal/imageutil"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func TestPlane(t *testing.T) {
	s := scene.NewScene(light.NewPoint(
		light.Intensity(1),
		light.Color(color.RGBA{0, 128, 255, 255}),
		light.Position(math.NewVec3[float32](2, 2, 2)),
	), model.NewPlane(1, 1))
	r := render.NewRenderer(
		render.Camera(camera.NewPerspective(
			camera.Position(math.NewVec3[float32](2, 2, 2)),
			camera.ViewFrustum(45, 1, 0.1, 10),
		)),
		render.Size(500, 500),
		render.MSAA(2),
		render.Scene(s),
		render.Background(color.FromHex("#181818")),
	)
	imageutil.Save(r.Render(), "plane.png")
}
