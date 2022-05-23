// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

// AABB an axis aligned bounding box
type AABB struct {
	Min, Max math.Vec3[float32]
}

// NewAABB computes a new axis aligned bounding box of given vertices
func NewAABB(vs ...math.Vec3[float32]) AABB {
	min := math.NewVec3[float32](math.MaxFloat32, math.MaxFloat32, math.MaxFloat32)
	max := math.NewVec3[float32](-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32)
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
	minX := math.Max(aabb.Min.X, aabb2.Min.X)
	minY := math.Max(aabb.Min.Y, aabb2.Min.Y)
	minZ := math.Max(aabb.Min.Z, aabb2.Min.Z)
	maxX := math.Min(aabb.Max.X, aabb2.Max.X)
	maxY := math.Min(aabb.Max.Y, aabb2.Max.Y)
	maxZ := math.Min(aabb.Max.Y, aabb2.Max.Z)

	return minX <= maxX && minY <= maxY && minZ <= maxZ
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
func (aabb *AABB) Eq(aabb2 AABB) bool {
	return aabb.Min.Eq(aabb2.Min) && aabb.Max.Eq(aabb2.Max)
}

// Contains checks if all given positions are inside the
// current bounding box, including points on the boundary.
func (aabb *AABB) Contains(vs ...math.Vec3[float32]) bool {
	lessEq := func(v1, v2 float32) bool {
		return math.ApproxEq(v1, v2, math.Epsilon) || math.ApproxLess(v1, v2, math.Epsilon)
	}
	lessEqVec := func(v1, v2 math.Vec3[float32]) bool {
		return lessEq(v1.X, v2.X) && lessEq(v1.Y, v2.Y) && lessEq(v1.Z, v2.Z)
	}

	for i := range vs {
		if !(lessEqVec(aabb.Min, vs[i]) && lessEqVec(vs[i], aabb.Max)) {
			return false
		}
	}
	return true
}
