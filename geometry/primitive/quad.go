// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

var _ Face = &Quad{}

// Quad is a quadrilateral that contains four vertices
type Quad struct {
	v1, v2, v3, v4 Vertex
	normal         math.Vec4
	aabb           *AABB
}

func NewQuad(v1, v2, v3, v4 *Vertex) *Quad {
	xmax := math.Max(v1.Pos.X, v2.Pos.X, v3.Pos.X, v4.Pos.X)
	xmin := math.Min(v1.Pos.X, v2.Pos.X, v3.Pos.X, v4.Pos.X)
	ymax := math.Max(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y, v4.Pos.Y)
	ymin := math.Min(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y, v4.Pos.Y)
	zmax := math.Max(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z, v4.Pos.Z)
	zmin := math.Min(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z, v4.Pos.Z)
	min := math.NewVec3(xmin, ymin, zmin)
	max := math.NewVec3(xmax, ymax, zmax)

	return &Quad{
		v1: *v1, v2: *v2, v3: *v3, v4: *v4, aabb: &AABB{min, max},
	}
}

func (q *Quad) AABB() AABB {
	if q.aabb == nil {
		xmax := math.Max(q.v1.Pos.X, q.v2.Pos.X, q.v3.Pos.X, q.v4.Pos.X)
		xmin := math.Min(q.v1.Pos.X, q.v2.Pos.X, q.v3.Pos.X, q.v4.Pos.X)
		ymax := math.Max(q.v1.Pos.Y, q.v2.Pos.Y, q.v3.Pos.Y, q.v4.Pos.Y)
		ymin := math.Min(q.v1.Pos.Y, q.v2.Pos.Y, q.v3.Pos.Y, q.v4.Pos.Y)
		zmax := math.Max(q.v1.Pos.Z, q.v2.Pos.Z, q.v3.Pos.Z, q.v4.Pos.Z)
		zmin := math.Min(q.v1.Pos.Z, q.v2.Pos.Z, q.v3.Pos.Z, q.v4.Pos.Z)
		min := math.NewVec3(xmin, ymin, zmin)
		max := math.NewVec3(xmax, ymax, zmax)
		q.aabb = &AABB{min, max}
	}
	return *q.aabb
}

func (q *Quad) Vertices(f func(v *Vertex) bool) {
	if !f(&q.v1) || !f(&q.v2) || !f(&q.v3) || !f(&q.v4) {
		return
	}
}

func (q *Quad) Triangles(f func(*Triangle) bool) {
	f(NewTriangle(&q.v1, &q.v2, &q.v3))
	f(NewTriangle(&q.v1, &q.v3, &q.v4))
}

func (q *Quad) Normal() math.Vec4 {
	return q.normal
}
