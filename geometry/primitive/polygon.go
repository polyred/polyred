// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"errors"

	"changkun.de/x/ddd/math"
)

// Polygon is a polygon that contains multiple vertices.
type Polygon struct {
	vs   []Vertex
	aabb *AABB
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
	min := math.NewVector(xmin, ymin, zmin, 1)
	max := math.NewVector(xmax, ymax, zmax, 1)
	p.aabb = &AABB{min, max}
	return p, nil
}
