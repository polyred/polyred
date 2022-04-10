// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"fmt"

	"poly.red/math"
)

var _ Face[float32] = &Triangle{}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	ID         uint64
	V1, V2, V3 *Vertex
	MaterialID uint64

	aabb       *AABB
	faceNormal math.Vec4[float32]
}

// NewTriangle creates a new triangle using the given three vertices.
// This method does not check the validity of the three vertices.
// Instead, one can check if the three vertices can verify a triangle
// using IsValid method.
//
// The returned Triangle has an unset index.
func NewTriangle(v1, v2, v3 *Vertex) *Triangle {
	xmax := math.Max(v1.Pos.X, v2.Pos.X, v3.Pos.X)
	xmin := math.Min(v1.Pos.X, v2.Pos.X, v3.Pos.X)
	ymax := math.Max(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y)
	ymin := math.Min(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y)
	zmax := math.Max(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z)
	zmin := math.Min(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z)
	min := math.NewVec3(xmin, ymin, zmin)
	max := math.NewVec3(xmax, ymax, zmax)
	v2v1 := v1.Pos.Sub(v2.Pos)
	v2v3 := v3.Pos.Sub(v2.Pos)

	return &Triangle{
		V1:         v1,
		V2:         v2,
		V3:         v3,
		faceNormal: v2v3.Cross(v2v1).Unit(),
		aabb:       &AABB{min, max},
	}
}

func (t *Triangle) String() string {
	return fmt.Sprintf(`&Triangle{
	ID: %d,
	V1: %+v,
	V2: %+v,
	V3: %+v,
	MatID: %d,
}`, t.ID, t.V1, t.V2, t.V3, t.MaterialID)
}

// IsValid is an assertion to check if the given triangle is valid or not.
func (t *Triangle) IsValid() bool {
	p1 := t.V1.Pos
	p2 := t.V2.Pos
	p3 := t.V3.Pos

	p1p2 := p2.Sub(p1)
	p1p3 := p3.Sub(p1)
	if p1p2.IsZero() {
		return false
	}
	if p1p3.IsZero() {
		return false
	}

	d := p1p2.Dot(p1p3) / (p1p2.Len() * p1p3.Len())
	return !math.ApproxEq(d, 1, math.Epsilon) &&
		!math.ApproxEq(d, -1, math.Epsilon)
}

// Area returns the surface area of the given triangle.
func (t *Triangle) Area() float32 {
	p1 := t.V1.Pos
	p2 := t.V2.Pos
	p3 := t.V3.Pos

	p1p2 := p2.Sub(p1)
	p1p3 := p3.Sub(p1)

	if p1p2.IsZero() {
		return 0
	}
	if p1p3.IsZero() {
		return 0
	}

	return 0.5 * p1p2.Cross(p1p3).Len()
}

// AABB returns the AABB of the given triangle.
func (t *Triangle) AABB() AABB {
	if t.aabb == nil {
		xmax := math.Max(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)
		xmin := math.Min(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)
		ymax := math.Max(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)
		ymin := math.Min(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)
		zmax := math.Max(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)
		zmin := math.Min(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)
		min := math.NewVec3(xmin, ymin, zmin)
		max := math.NewVec3(xmax, ymax, zmax)
		t.aabb = &AABB{min, max}
	}

	return *t.aabb
}

// Vertices traserval all vertices of the given triangle.
func (t *Triangle) Vertices(f func(v *Vertex) bool) {
	if !f(t.V1) || !f(t.V2) || !f(t.V3) {
		return
	}
}

// Triangles traversal all triangles of the given triangle.
func (t *Triangle) Triangles(f func(*Triangle) bool) {
	f(t)
}

// Normal returns the face normal of the given triangle.
func (t *Triangle) Normal() math.Vec4[float32] {
	if t.faceNormal.IsZero() {
		v2v1 := t.V1.Pos.Sub(t.V2.Pos)
		v2v3 := t.V3.Pos.Sub(t.V2.Pos)
		t.faceNormal = v2v3.Cross(v2v1).Unit()
	}

	return t.faceNormal
}
