// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"image/color"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/math"
)

// NewPlane returns a triangle soup that represents a plane with the
// given width and height.
func NewPlane(width, height float64) Mesh {
	v1 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec4(0, 1, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{255, 0, 0, 255},
	}
	v2 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec4(0, 0, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 255, 0, 255},
	}
	v3 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec4(1, 0, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 255, 255},
	}
	v4 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec4(1, 1, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 0, 255},
	}
	return NewTriangleSoup([]*primitive.Triangle{
		{V1: v1, V2: v2, V3: v3},
		{V1: v1, V2: v3, V3: v4},
	})
}
