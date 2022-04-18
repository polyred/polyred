// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene_test

import (
	"testing"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

func TestGroup_Normalize(t *testing.T) {
	s := scene.NewScene(geometry.NewWith(model.NewPlane(1, 1), nil))
	r := render.NewRenderer(
		render.Camera(camera.NewOrthographic(
			camera.Position(math.NewVec3[float32](0, 1, 0)),
			camera.LookAt(math.NewVec3[float32](0, 0, 0),
				math.NewVec3[float32](0, 0, -1)),
			camera.ViewFrustum(-1, 1, -1, 1, 1, -1),
		)),
		render.Size(500, 500),
		render.MSAA(2),
		render.Scene(s),
		render.Background(color.FromHex("#181818")),
	)
	s.Normalize()
	imageutil.Save(r.Render(), "../internal/examples/out/normalize.png")
}
