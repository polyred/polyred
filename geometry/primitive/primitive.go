// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Pos math.Vector
	UV  math.Vector
	Nor math.Vector
	Col color.RGBA
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos,
		Max: v.Pos,
	}
}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	V1, V2, V3 Vertex
	FaceNormal math.Vector
}

func (t *Triangle) AABB() AABB {
	xMax := math.Max(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)
	xMin := math.Min(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)

	yMax := math.Max(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)
	yMin := math.Min(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)

	zMax := math.Max(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)
	zMin := math.Min(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)

	return AABB{
		Min: math.NewVector(xMin, yMin, zMin, 1),
		Max: math.NewVector(xMax, yMax, zMax, 1),
	}
}
