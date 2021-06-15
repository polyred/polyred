// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	_ "image/jpeg" // for jpg encoding

	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
)

// TriangleMesh implements a triangular mesh.
type TriangleMesh struct {
	Faces    []*primitive.Triangle
	Material material.Material

	// aabb must be transformed when applying the context.
	aabb AABB
	// context is a transformation context (model matrix) that accumulates
	// applied transformation matrices (multiplied from left side) for the
	// given mesh.
	//
	// context is a persistant status for the given mesh and can be reused
	// for each of the rendering frame unless the mesh intentionally calls
	// resetContext() method.
	context math.Matrix
}

// NewTriangleMesh returns a triangular mesh.
func NewTriangleMesh(ts []*primitive.Triangle) *TriangleMesh {
	// Compute AABB at loading time.
	min := math.Vector{X: math.MaxFloat64, Y: math.MaxFloat64, Z: math.MaxFloat64, W: 1}
	max := math.Vector{X: -math.MaxFloat64, Y: -math.MaxFloat64, Z: -math.MaxFloat64, W: 1}
	for i := 0; i < len(ts); i++ {
		min.X = math.Min(min.X, ts[i].V1.Pos.X, ts[i].V2.Pos.X, ts[i].V3.Pos.X)
		min.Y = math.Min(min.Y, ts[i].V1.Pos.Y, ts[i].V2.Pos.Y, ts[i].V3.Pos.Y)
		min.Z = math.Min(min.Z, ts[i].V1.Pos.Z, ts[i].V2.Pos.Z, ts[i].V3.Pos.Z)
		max.X = math.Max(min.X, ts[i].V1.Pos.X, ts[i].V2.Pos.X, ts[i].V3.Pos.X)
		max.Y = math.Max(min.Y, ts[i].V1.Pos.Y, ts[i].V2.Pos.Y, ts[i].V3.Pos.Y)
		max.Z = math.Max(min.Z, ts[i].V1.Pos.Z, ts[i].V2.Pos.Z, ts[i].V3.Pos.Z)
	}

	return &TriangleMesh{
		Faces:   ts,
		aabb:    AABB{min, max},
		context: math.MatI,
	}
}

func (t *TriangleMesh) UseMaterial(mat material.Material) {
	t.Material = mat
}

// modelMatrix returns the transformation context as the model matrix
// for the current frame (or at call time).
func (t *TriangleMesh) ModelMatrix() math.Matrix {
	return t.context
}

// normalMatrix can be ((Tcamera * Tmodel)^(-1))^T or ((Tmodel)^(-1))^T
// depending on which transformation space. Here we use the 2nd form,
// i.e. model space normal matrix to save some computation of camera
// transforamtion in the shading process.
// The reason we need normal matrix is that normals are transformed
// incorrectly using MVP matrices. However, a normal matrix helps us
// to fix the problem.
func (t *TriangleMesh) NormalMatrix() math.Matrix {
	return t.ModelMatrix().Inv().T()
}

func (t *TriangleMesh) ResetContext() {
	t.context = math.MatI
}

// Scale sets the scale matrix.
func (m *TriangleMesh) Scale(sx, sy, sz float64) {
	m.context = math.NewMatrix(
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(m.context)
}

// SetTranslate sets the translate matrix.
func (m *TriangleMesh) Translate(tx, ty, tz float64) {
	m.context = math.NewMatrix(
		1, 0, 0, tx,
		0, 1, 0, ty,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(m.context)
}

func (m *TriangleMesh) Rotate(dir math.Vector, angle float64) {
	u := dir.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	m.context = q.ToRoMat().MulM(m.context)
}

func (m *TriangleMesh) AABB() AABB {
	min := m.aabb.Min.Apply(m.ModelMatrix())
	max := m.aabb.Max.Apply(m.ModelMatrix())
	return AABB{min, max}
}

func (m *TriangleMesh) Center() math.Vector {
	aabb := m.AABB()
	return aabb.Min.Add(aabb.Max).Pos()
}

func (m *TriangleMesh) Radius() float64 {
	aabb := m.AABB()
	return aabb.Max.Sub(aabb.Min).Len() / 2
}

// Normalize rescales the mesh to the unit sphere centered at the origin.
func (m *TriangleMesh) Normalize() {
	aabb := m.AABB()
	center := aabb.Min.Add(aabb.Max).Pos()
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	for i := 0; i < len(m.Faces); i++ {
		f := m.Faces[i]
		f.V1.Pos = f.V1.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		f.V2.Pos = f.V2.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		f.V3.Pos = f.V3.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	}

	// update AABB after scaling
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	m.aabb = AABB{min, max}
	m.ResetContext()
}
