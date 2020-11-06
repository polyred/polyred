// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import "math"

// AABB an axis aligned bounding box
type AABB struct {
	min, max Vector
}

// NewAABB computes a new axis aligned bounding box of given vertices
func NewAABB(v1, v2, v3 Vertex) AABB {
	xMax := math.Max(math.Max(v1.Position.X, v2.Position.X), v3.Position.X)
	xMin := math.Min(math.Min(v1.Position.X, v2.Position.X), v3.Position.X)

	yMax := math.Max(math.Max(v1.Position.Y, v2.Position.Y), v3.Position.Y)
	yMin := math.Min(math.Min(v1.Position.Y, v2.Position.Y), v3.Position.Y)

	zMax := math.Max(math.Max(v1.Position.Z, v2.Position.Z), v3.Position.Z)
	zMin := math.Min(math.Min(v1.Position.Z, v2.Position.Z), v3.Position.Z)

	return AABB{
		min: Vector{xMin, yMin, zMin, 1},
		max: Vector{xMax, yMax, zMax, 1},
	}
}
