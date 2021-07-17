// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene_test

import (
	"fmt"
	"testing"

	"poly.red/geometry"
	"poly.red/math"
	"poly.red/object"
	"poly.red/scene"
)

func TestScene(t *testing.T) {
	s := scene.NewScene()
	p1 := geometry.NewPlane(1, 1)
	p2 := geometry.NewPlane(1, 2)
	p3 := geometry.NewPlane(2, 1)
	p4 := geometry.NewPlane(2, 2)
	p5 := geometry.NewPlane(3, 2)
	p6 := geometry.NewPlane(2, 3)
	p7 := geometry.NewPlane(3, 3)

	g1 := s.Add(p1)
	g2 := g1.Add(p2, p3)
	g3 := g2.Add(p4)
	g3.Add(p5, p6, p7)

	g1.Scale(2, 2, 2)
	g1.Translate(1, 2, 3)

	s.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
		fmt.Println(o, modelMatrix)
		return true
	})
}
