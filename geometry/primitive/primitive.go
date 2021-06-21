// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"
	"math/rand"

	"changkun.de/x/ddd/math"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Pos math.Vector
	UV  math.Vector
	Nor math.Vector
	Col color.RGBA
}

func NewRandomVertex() *Vertex {
	return &Vertex{
		Pos: math.NewVector(rand.Float64(), rand.Float64(), rand.Float64(), 1),
		UV:  math.NewVector(rand.Float64(), rand.Float64(), 0, 1),
		Nor: math.NewVector(rand.Float64(), rand.Float64(), rand.Float64(), 1).Unit(),
		Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())},
	}
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos,
		Max: v.Pos,
	}
}

type Face interface {
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}

var (
	_ Face = &Triangle{}
)

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	V1, V2, V3 Vertex

	faceNormal math.Vector
	aabb       *AABB
}

func NewTriangle(v1, v2, v3 *Vertex) *Triangle {
	xmax := math.Max(v1.Pos.X, v2.Pos.X, v3.Pos.X)
	xmin := math.Min(v1.Pos.X, v2.Pos.X, v3.Pos.X)
	ymax := math.Max(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y)
	ymin := math.Min(v1.Pos.Y, v2.Pos.Y, v3.Pos.Y)
	zmax := math.Max(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z)
	zmin := math.Min(v1.Pos.Z, v2.Pos.Z, v3.Pos.Z)
	min := math.NewVector(xmin, ymin, zmin, 1)
	max := math.NewVector(xmax, ymax, zmax, 1)
	v2v1 := v1.Pos.Sub(v2.Pos)
	v2v3 := v3.Pos.Sub(v2.Pos)

	return &Triangle{
		V1:         *v1,
		V2:         *v2,
		V3:         *v3,
		faceNormal: v2v3.Cross(v2v1).Unit(),
		aabb:       &AABB{min, max},
	}
}

func (t *Triangle) AABB() AABB {
	if t.aabb == nil {
		xmax := math.Max(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)
		xmin := math.Min(t.V1.Pos.X, t.V2.Pos.X, t.V3.Pos.X)
		ymax := math.Max(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)
		ymin := math.Min(t.V1.Pos.Y, t.V2.Pos.Y, t.V3.Pos.Y)
		zmax := math.Max(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)
		zmin := math.Min(t.V1.Pos.Z, t.V2.Pos.Z, t.V3.Pos.Z)
		min := math.NewVector(xmin, ymin, zmin, 1)
		max := math.NewVector(xmax, ymax, zmax, 1)
		t.aabb = &AABB{min, max}
	}

	return *t.aabb
}

func (t *Triangle) Vertices(f func(v *Vertex) bool) {
	if !f(&t.V1) || !f(&t.V2) || !f(&t.V3) {
		return
	}
}

func (t *Triangle) Triangles(f func(*Triangle) bool) {
	f(t)
}

func (t *Triangle) FaceNormal() math.Vector {
	if t.faceNormal.IsZero() {
		v2v1 := t.V1.Pos.Sub(t.V2.Pos)
		v2v3 := t.V3.Pos.Sub(t.V2.Pos)
		t.faceNormal = v2v3.Cross(v2v1).Unit()
	}

	return t.faceNormal
}
