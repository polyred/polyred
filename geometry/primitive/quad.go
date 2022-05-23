// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

var _ Face = &Quad{}

// Quad is a quadrilateral that contains four vertices
type Quad struct {
	ID             uint64
	V1, V2, V3, V4 *Vertex
	MaterialID     int64

	normal math.Vec4[float32]
	aabb   *AABB
}

func NewQuad[T math.Float](v1, v2, v3, v4 *Vertex) *Quad {
	xmax := math.Max(v1.Pos.X, v2.Pos.X, v3.Pos.X, v4.Pos.X)
	xmin := math.Min(v1.Pos.X, v2.Pos.X, v3.Pos.X, v4.Pos.X)
	ymax := math.Max(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y, v4.Pos.Y)
	ymin := math.Min(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y, v4.Pos.Y)
	zmax := math.Max(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z, v4.Pos.Z)
	zmin := math.Min(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z, v4.Pos.Z)
	min := math.NewVec3(xmin, ymin, zmin)
	max := math.NewVec3(xmax, ymax, zmax)

	return &Quad{V1: v1, V2: v2, V3: v3, V4: v4, aabb: &AABB{min, max}}
}

func (q *Quad) AABB() AABB {
	if q.aabb == nil {
		xmax := math.Max(q.V1.Pos.X, q.V2.Pos.X, q.V3.Pos.X, q.V4.Pos.X)
		xmin := math.Min(q.V1.Pos.X, q.V2.Pos.X, q.V3.Pos.X, q.V4.Pos.X)
		ymax := math.Max(q.V1.Pos.Y, q.V2.Pos.Y, q.V3.Pos.Y, q.V4.Pos.Y)
		ymin := math.Min(q.V1.Pos.Y, q.V2.Pos.Y, q.V3.Pos.Y, q.V4.Pos.Y)
		zmax := math.Max(q.V1.Pos.Z, q.V2.Pos.Z, q.V3.Pos.Z, q.V4.Pos.Z)
		zmin := math.Min(q.V1.Pos.Z, q.V2.Pos.Z, q.V3.Pos.Z, q.V4.Pos.Z)
		min := math.NewVec3(xmin, ymin, zmin)
		max := math.NewVec3(xmax, ymax, zmax)
		q.aabb = &AABB{min, max}
	}
	return *q.aabb
}

func (q *Quad) Vertices(f func(v *Vertex) bool) {
	if !f(q.V1) || !f(q.V2) || !f(q.V3) || !f(q.V4) {
		return
	}
}

func (q *Quad) Triangles(f func(*Triangle) bool) {
	t := NewTriangle(q.V1, q.V2, q.V3)
	t.MaterialID = q.MaterialID
	f(t)

	t = NewTriangle(q.V1, q.V3, q.V4)
	t.MaterialID = q.MaterialID
	f(t)
}

func (q *Quad) Normal() math.Vec4[float32] {
	return q.normal
}
