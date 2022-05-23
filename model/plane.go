// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model

import (
	"image/color"

	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

// NewPlane returns a triangle soup that represents a plane with the
// given width and height.
func NewPlane(width, height float32) mesh.Mesh {
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
	return mesh.NewPolygonMesh([]primitive.Face{
		&primitive.Triangle{V1: v1, V2: v2, V3: v3, MaterialID: -1},
		&primitive.Triangle{V1: v1, V2: v3, V3: v4, MaterialID: -1},
	})
}
