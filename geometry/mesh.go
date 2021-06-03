// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package geometry

import (
	_ "image/jpeg" // for jpg encoding

	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Position math.Vector
	Color    math.Vector
	UV       math.Vector
	Normal   math.Vector
}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	V1, V2, V3 Vertex
}

// TriangleMesh implements a triangular mesh.
type TriangleMesh struct {
	Faces    []*Triangle
	Material material.Material

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
func NewTriangleMesh(ts []*Triangle) *TriangleMesh {
	return &TriangleMesh{
		Faces:   ts,
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
