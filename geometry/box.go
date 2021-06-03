// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package geometry

import "changkun.de/x/ddd/math"

// AABB an axis aligned bounding box
type AABB struct {
	Min, Max math.Vector
}

// NewAABB computes a new axis aligned bounding box of given vertices
func NewAABB(v1, v2, v3 math.Vector) AABB {
	xMax := math.Max(v1.X, v2.X, v3.X)
	xMin := math.Min(v1.X, v2.X, v3.X)

	yMax := math.Max(v1.Y, v2.Y, v3.Y)
	yMin := math.Min(v1.Y, v2.Y, v3.Y)

	zMax := math.Max(v1.Z, v2.Z, v3.Z)
	zMin := math.Min(v1.Z, v2.Z, v3.Z)

	return AABB{
		Min: math.NewVector(xMin, yMin, zMin, 1),
		Max: math.NewVector(xMax, yMax, zMax, 1),
	}
}

// Intersect checks if the two given AABBs share an intersection.
// If the two AABBs only share a single vertex or a 2D plane, then
// it is also considered as an intersection and returns true.
func (aabb AABB) Intersect(aabb2 AABB) bool {
	min := math.NewVector(
		math.Max(aabb.Min.X, aabb2.Min.X),
		math.Max(aabb.Min.Y, aabb2.Min.Y),
		math.Max(aabb.Min.Z, aabb2.Min.Z),
		1,
	)
	max := math.NewVector(
		math.Min(aabb.Max.X, aabb2.Max.X),
		math.Min(aabb.Max.Y, aabb2.Max.Y),
		math.Min(aabb.Max.Y, aabb2.Max.Z),
		1,
	)

	if min.X <= max.X && min.Y <= max.Y && min.Z <= max.Z {
		return true
	}
	return false
}
