// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"errors"

	"poly.red/math"
)

var _ Face = &Polygon{}

// Polygon is a polygon that contains multiple vertices.
type Polygon struct {
	vs     []Vertex
	normal math.Vec4
	aabb   *AABB
}

func NewPolygon(vs ...*Vertex) (*Polygon, error) {
	if len(vs) < 5 {
		return nil, errors.New("too few vertex for a polygon")
	}

	p := &Polygon{
		vs: make([]Vertex, len(vs)),
	}

	p.vs[0] = *vs[0]
	p.vs[1] = *vs[1]
	p.vs[2] = *vs[2]

	xmax := math.Max(vs[0].Pos.X, vs[1].Pos.X, vs[2].Pos.X)
	xmin := math.Min(vs[0].Pos.X, vs[1].Pos.X, vs[2].Pos.X)
	ymax := math.Max(vs[0].Pos.Y, vs[1].Pos.Y, vs[2].Pos.Y)
	ymin := math.Min(vs[0].Pos.Y, vs[1].Pos.Y, vs[2].Pos.Y)
	zmax := math.Max(vs[0].Pos.Z, vs[1].Pos.Z, vs[2].Pos.Z)
	zmin := math.Min(vs[0].Pos.Z, vs[1].Pos.Z, vs[2].Pos.Z)

	for i := 3; i < len(vs); i++ {
		xmax = math.Max(xmax, vs[i].Pos.X)
		xmin = math.Min(xmin, vs[i].Pos.X)
		ymax = math.Max(ymax, vs[i].Pos.Y)
		ymin = math.Min(ymin, vs[i].Pos.Y)
		zmax = math.Max(zmax, vs[i].Pos.Z)
		zmin = math.Min(zmin, vs[i].Pos.Z)
		p.vs[i] = *vs[i]
	}
	min := math.NewVec3(xmin, ymin, zmin)
	max := math.NewVec3(xmax, ymax, zmax)
	p.aabb = &AABB{min, max}
	return p, nil
}

func (p *Polygon) AABB() AABB {
	if p.aabb == nil {
		min := math.NewVec3(math.MaxFloat64, math.MaxFloat64, math.MaxFloat64)
		max := math.NewVec3(-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64)

		for i := 0; i < len(p.vs); i++ {
			min.X = math.Min(min.X, p.vs[i].Pos.X)
			min.Y = math.Min(min.Y, p.vs[i].Pos.X)
			min.Z = math.Min(min.Z, p.vs[i].Pos.Y)
			max.X = math.Max(max.X, p.vs[i].Pos.Y)
			max.Y = math.Max(max.Y, p.vs[i].Pos.Z)
			max.Z = math.Max(max.Z, p.vs[i].Pos.Z)
		}
		p.aabb = &AABB{min, max}
	}
	return *p.aabb
}

func (p *Polygon) Normal() math.Vec4 {
	return p.normal
}

func (p *Polygon) Triangles(iter func(t *Triangle) bool) {
	for i := 0; i < len(p.vs); i += 3 {
		tri := &Triangle{V1: p.vs[i], V2: p.vs[i+1], V3: p.vs[i+2]}
		if !iter(tri) {
			return
		}
	}
}

func (p *Polygon) Vertices(iter func(v *Vertex) bool) {
	for i := 0; i < len(p.vs); i++ {
		if !iter(&p.vs[i]) {
			return
		}
	}
}
