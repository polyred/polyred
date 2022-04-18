// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"poly.red/buffer"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ object.Object[float32] = &Geometry{}
	_ mesh.Mesh[float32]     = &Geometry{} // FIXME: geometry should or should not implements mesh (?)
)

// Geometry represents a geometric object that can be rendered.
// A geometry consists of a vertex-based object and a list of materials.
// The vertices of the object contains an ID that refers the to associated list of materials.
type Geometry struct {
	math.TransformContext[float32]

	mesh mesh.Mesh[float32]
	mats []material.ID
}

func New(mesh mesh.Mesh[float32], ids ...material.ID) *Geometry {
	g := &Geometry{
		mesh: mesh,
		mats: ids,
	}
	g.ResetContext()

	// FIXME: If a given mesh have no materials for its primitives, what should we do?

	return g
}

func (g *Geometry) Materials() []material.ID {
	return g.mats
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
