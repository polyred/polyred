// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive_test

import (
	"image/color"
	"log"
	"testing"

	"poly.red/geometry/primitive"
	"poly.red/math"
)

func TestNewPolygon(t *testing.T) {
	width, height := float32(0.5), float32(0.5)
	v1 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(-0.5*width, 0, -0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](0, 1)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{255, 0, 0, 255}),
	)
	v2 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(-0.5*width, 0, 0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](0, 0)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 255, 0, 255}),
	)
	v3 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(0.5*width, 0, 0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](1, 0)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 0, 255, 255}),
	)
	v4 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(0.5*width, 0, -0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](1, 1)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 0, 0, 255}),
	)

	poly := primitive.NewPolygon(v1, v2, v3, v4)
	poly.Triangles(func(t *primitive.Triangle) bool {
		log.Println(t)
		return true
	})
}
