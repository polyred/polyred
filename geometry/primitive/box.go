// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "changkun.de/x/polyred/math"

// AABB an axis aligned bounding box
type AABB struct {
	Min, Max math.Vec3
}

// NewAABB computes a new axis aligned bounding box of given vertices
func NewAABB(vs ...math.Vec3) AABB {
	min := math.NewVec3(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
	max := math.NewVec3(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)
	for i := range vs {
		min.X = math.Min(min.X, vs[i].X)
		min.Y = math.Min(min.Y, vs[i].Y)
		min.Z = math.Min(min.Z, vs[i].Z)
		max.X = math.Max(max.X, vs[i].X)
		max.Y = math.Max(max.Y, vs[i].Y)
		max.Z = math.Max(max.Z, vs[i].Z)
	}
	return AABB{min, max}
}

// Intersect checks if the two given AABBs share an intersection.
// If the two AABBs only share a single vertex or a 2D plane, then
// it is also considered as an intersection and returns true.
func (aabb *AABB) Intersect(aabb2 AABB) bool {
	min := math.NewVec4(
		math.Max(aabb.Min.X, aabb2.Min.X),
		math.Max(aabb.Min.Y, aabb2.Min.Y),
		math.Max(aabb.Min.Z, aabb2.Min.Z),
		1,
	)
	max := math.NewVec4(
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

// Add adds a given aabb to the current aabb
func (aabb *AABB) Add(aabb2 AABB) {
	aabb.Min.X = math.Min(aabb.Min.X, aabb2.Min.X)
	aabb.Min.Y = math.Min(aabb.Min.Y, aabb2.Min.Y)
	aabb.Min.Z = math.Min(aabb.Min.Z, aabb2.Min.Z)

	aabb.Max.X = math.Max(aabb.Max.X, aabb2.Max.X)
	aabb.Max.Y = math.Max(aabb.Max.Y, aabb2.Max.Y)
	aabb.Max.Z = math.Max(aabb.Max.Z, aabb2.Max.Z)
}

// Eq checks if two aabbs are equal
func (aabb AABB) Eq(aabb2 AABB) bool {
	if aabb.Min.Eq(aabb2.Min) && aabb.Max.Eq(aabb2.Max) {
		return true
	}
	return false
}
