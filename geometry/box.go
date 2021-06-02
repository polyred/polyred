// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package geometry

import "changkun.de/x/ddd/math"

// AABB an axis aligned bounding box
type AABB struct {
	min, max math.Vector
}

// NewAABB computes a new axis aligned bounding box of given vertices
func NewAABB(v1, v2, v3 Vertex) AABB {
	xMax := math.Max(v1.Position.X, v2.Position.X, v3.Position.X)
	xMin := math.Min(v1.Position.X, v2.Position.X, v3.Position.X)

	yMax := math.Max(v1.Position.Y, v2.Position.Y, v3.Position.Y)
	yMin := math.Min(v1.Position.Y, v2.Position.Y, v3.Position.Y)

	zMax := math.Max(v1.Position.Z, v2.Position.Z, v3.Position.Z)
	zMin := math.Min(v1.Position.Z, v2.Position.Z, v3.Position.Z)

	return AABB{
		min: math.NewVector(xMin, yMin, zMin, 1),
		max: math.NewVector(xMax, yMax, zMax, 1),
	}
}

// Intersect checks if the two given AABBs share an intersection.
// If the two AABBs only share a single vertex or a 2D plane, then
// it is also considered as an intersection and returns true.
func (aabb AABB) Intersect(aabb2 AABB) bool {
	min := math.NewVector(
		math.Max(aabb.min.X, aabb2.min.X),
		math.Max(aabb.min.X, aabb2.min.X),
		math.Max(aabb.min.X, aabb2.min.X),
		1,
	)
	max := math.NewVector(
		math.Min(aabb.min.X, aabb2.min.X),
		math.Min(aabb.min.X, aabb2.min.X),
		math.Min(aabb.min.X, aabb2.min.X),
		1,
	)

	if min.X <= max.X && min.Y <= max.Y && min.Z <= max.Z {
		return true
	}
	return false
}
