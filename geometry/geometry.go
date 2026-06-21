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
// A geometry consists of a vertex-based object and the materials it owns. Each
// primitive's MaterialID is a geometry-local index into this list (or negative
// to use vertex color); the renderer tabulates these per frame. There is no
// global material pool.
type Geometry struct {
	mesh mesh.Mesh
	mats []material.Material

	math.TransformContext[float32]
}

// New builds a geometry from a mesh and the materials it owns. The mesh's
// primitives carry geometry-local indices (0..len(mats)-1) into mats.
func New(mesh mesh.Mesh, mats ...material.Material) *Geometry {
	g := &Geometry{
		mesh: mesh,
		mats: mats,
	}
	g.ResetContext()
	return g
}

// Materials returns the materials this geometry owns, indexed by the local
// MaterialID carried on its primitives.
func (g *Geometry) Materials() []material.Material {
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
