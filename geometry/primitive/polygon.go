// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"poly.red/math"
)

var _ Face = &Polygon{}

// Polygon is a polygon that contains multiple vertices.
type Polygon struct {
	Verts      []*Vertex
	MaterialID int64

	aabb *AABB
}

func NewPolygon(vs ...*Vertex) *Polygon {
	p := &Polygon{
		Verts:      make([]*Vertex, len(vs)),
		MaterialID: -1,
	}

	p.Verts[0] = vs[0]
	p.Verts[1] = vs[1]
	p.Verts[2] = vs[2]

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
		p.Verts[i] = vs[i]
	}
	min := math.NewVec3(xmin, ymin, zmin)
	max := math.NewVec3(xmax, ymax, zmax)
	p.aabb = &AABB{min, max}
	return p
}

func (p *Polygon) AABB() AABB {
	if p.aabb == nil {
		min := math.NewVec3[float32](math.MaxFloat32, math.MaxFloat32, math.MaxFloat32)
		max := math.NewVec3[float32](-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32)

		for i := 0; i < len(p.Verts); i++ {
			min.X = math.Min(min.X, p.Verts[i].Pos.X)
			min.Y = math.Min(min.Y, p.Verts[i].Pos.X)
			min.Z = math.Min(min.Z, p.Verts[i].Pos.Y)
			max.X = math.Max(max.X, p.Verts[i].Pos.Y)
			max.Y = math.Max(max.Y, p.Verts[i].Pos.Z)
			max.Z = math.Max(max.Z, p.Verts[i].Pos.Z)
		}
		p.aabb = &AABB{min, max}
	}
	return *p.aabb
}

func (p *Polygon) Triangles(iter func(t *Triangle) bool) {
	for i := 0; i < len(p.Verts)-2; i++ {
		tri := &Triangle{V1: p.Verts[0], V2: p.Verts[i+1], V3: p.Verts[i+2], MaterialID: p.MaterialID}
		if !iter(tri) {
			return
		}
	}
}

func (p *Polygon) Vertices(iter func(v *Vertex) bool) {
	for i := 0; i < len(p.Verts); i++ {
		if !iter(p.Verts[i]) {
			return
		}
	}
}
