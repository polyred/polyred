// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"poly.red/buffer"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/internal/cache"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ object.Object[float32] = &Geometry{}
	_ mesh.Mesh[float32]     = &Geometry{}
)

type Geometry struct {
	math.TransformContext[float32]

	mesh     mesh.Mesh[float32]
	material material.Material
}

func New() *Geometry {
	g := &Geometry{}
	g.ResetContext()
	return g
}

func NewWith(mesh mesh.Mesh[float32], material material.Material) *Geometry {
	g := &Geometry{
		mesh:     mesh,
		material: material,
	}
	if material != nil {
		for _, t := range mesh.Triangles() {
			t.MaterialID = material.ID()
			cache.Set(t.MaterialID, material)
		}
	}
	g.ResetContext()
	return g
}

func (g *Geometry) SetMesh(m mesh.Mesh[float32]) {
	g.mesh = m
}

func (g *Geometry) SetMaterial(m material.Material) {
	g.material = m
	if m != nil {
		for _, t := range g.mesh.Triangles() {
			t.MaterialID = m.ID()
			cache.Set(t.MaterialID, m)
		}
	}
}

func (g *Geometry) AABB() primitive.AABB {
	return g.mesh.AABB()
}

func (g *Geometry) Name() string      { return "geometry" }
func (g *Geometry) Type() object.Type { return object.TypeGeometry }

func (g *Geometry) Triangles() []*primitive.Triangle {
	return g.mesh.Triangles()
}

func (g *Geometry) IndexBuffer() buffer.IndexBuffer {
	return g.mesh.IndexBuffer()
}

func (g *Geometry) VertexBuffer() buffer.VertexBuffer {
	return g.mesh.VertexBuffer()
}
