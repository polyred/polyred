// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"poly.red/math"
)

var _ Face = &Triangle{}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	V [3]Vertex

	faceNormal math.Vec3
	aabb       *AABB
}

// NewTriangle creates a new triangle using the given three vertices.
// This method does not check the validity of the three vertices.
// Instead, one can check if the three vertices can construct a triangle
// using IsValid method.
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
		V:          [3]Vertex{*v1, *v2, *v3},
		faceNormal: v2v3.Cross(v2v1).Unit().ToVec3(),
		aabb:       &AABB{min, max},
	}
}

// IsValid is an assertion to check if the given triangle is valid or not.
func (t *Triangle) IsValid() bool {
	p1 := t.V[0].Pos
	p2 := t.V[1].Pos
	p3 := t.V[2].Pos

	p1p2 := p2.Sub(p1)
	p1p3 := p3.Sub(p1)
	if p1p2.IsZero() {
		return false
	}
	if p1p3.IsZero() {
		return false
	}

	d := p1p2.Dot(p1p3) / (p1p2.Len() * p1p3.Len())
	if math.ApproxEq(d, 1, math.Epsilon) ||
		math.ApproxEq(d, -1, math.Epsilon) {
		return false
	}
	return true
}

// Area returns the surface area of the given triangle.
func (t *Triangle) Area() float64 {
	p1 := t.V[0].Pos
	p2 := t.V[1].Pos
	p3 := t.V[2].Pos

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
		xmax := math.Max(t.V[0].Pos.X, t.V[1].Pos.X, t.V[2].Pos.X)
		xmin := math.Min(t.V[0].Pos.X, t.V[1].Pos.X, t.V[2].Pos.X)
		ymax := math.Max(t.V[0].Pos.Y, t.V[1].Pos.Y, t.V[2].Pos.Y)
		ymin := math.Min(t.V[0].Pos.Y, t.V[1].Pos.Y, t.V[2].Pos.Y)
		zmax := math.Max(t.V[0].Pos.Z, t.V[1].Pos.Z, t.V[2].Pos.Z)
		zmin := math.Min(t.V[0].Pos.Z, t.V[1].Pos.Z, t.V[2].Pos.Z)
		min := math.NewVec3(xmin, ymin, zmin)
		max := math.NewVec3(xmax, ymax, zmax)
		t.aabb = &AABB{min, max}
	}

	return *t.aabb
}

// Vertices traserval all vertices of the given triangle.
func (t *Triangle) Vertices(f func(v *Vertex) bool) {
	if !f(&t.V[0]) || !f(&t.V[1]) || !f(&t.V[2]) {
		return
	}
}

// Triangles traversal all triangles of the given triangle.
func (t *Triangle) Triangles(f func(*Triangle) bool) {
	f(t)
}

// Normal returns the face normal of the given triangle.
func (t *Triangle) Normal() math.Vec3 {
	if t.faceNormal.IsZero() {
		v2v1 := t.V[0].Pos.Sub(t.V[1].Pos)
		v2v3 := t.V[2].Pos.Sub(t.V[1].Pos)
		t.faceNormal = v2v3.Cross(v2v1).Unit().ToVec3()
	}

	return t.faceNormal
}
