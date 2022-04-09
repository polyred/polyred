// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene_test

import (
	"testing"

	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
	"poly.red/scene/object"
)

func TestScene(t *testing.T) {
	s := scene.NewScene()
	p1 := model.NewPlane(1, 1)
	p2 := model.NewPlane(1, 2)
	p3 := model.NewPlane(2, 1)
	p4 := model.NewPlane(2, 2)
	p5 := model.NewPlane(3, 2)
	p6 := model.NewPlane(2, 3)
	p7 := model.NewPlane(3, 3)

	g1 := s.Add(p1)
	g2 := g1.Add(p2, p3)
	g3 := g2.Add(p4)
	g3.Add(p5, p6, p7)

	g1.Scale(2, 2, 2)
	g1.Translate(1, 2, 3)

	scene.IterObjects(s, func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		t.Log(o, modelMatrix)
		return true
	})
}
