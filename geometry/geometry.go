// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ object.Object[float32] = &Geometry{}
	// FIXME: geometry should or should not implements mesh (?)
	_ mesh.Mesh = &Geometry{}
)

// Geometry represents a geometric object that can be rendered.
// A geometry consists of a vertex-based object and a list of materials.
// The vertices of the object contains an ID that refers the to associated list of materials.
type Geometry struct {
	mesh mesh.Mesh
	mats []material.ID

	math.TransformContext[float32]
}

func New(mesh mesh.Mesh, ids ...material.ID) *Geometry {
	g := &Geometry{
		mesh: mesh,
		mats: ids,
	}
	g.ResetContext()

	// If we have multiple material IDs, let's don't do anything so far.
	if len(ids) != 1 {
		return g
	}

	// If there is only a single material ID, let assign them to all
	// primitives.
	// FIXME: todo

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
