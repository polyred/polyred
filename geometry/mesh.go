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

type Mesh interface {
	Rotate(r math.Vector, a float64)
	RotateX(a float64)
	RotateY(a float64)
	RotateZ(a float64)
	Translate(x, y, z float64)
	Scale(x, y, z float64)
	AABB() primitive.AABB
	Normalize()

	UseMaterial(m material.Material)
	NumTriangles() uint64
	Faces(func(f primitive.Face, m material.Material) bool)
	GetMaterial() material.Material
	ModelMatrix() math.Matrix
}

func NewBufferedMeshFromTriangleSoup(m *TriangleSoup) {

}

var (
	_ Mesh = &TriangleSoup{}
)

// TriangleSoup implements a triangular mesh.
type TriangleSoup struct {
	// all faces of the triangle mesh.
	faces []*primitive.Triangle
	// the corresponding material of the triangle mesh.
	material material.Material
	// aabb must be transformed when applying the context.
	aabb primitive.AABB
	// context is a transformation context (model matrix) that accumulates
	// applied transformation matrices (multiplied from left side) for the
	// given mesh.
	// context is a persistant status for the given mesh and can be reused
	// for each of the rendering frame unless the mesh intentionally calls
	// ResetContext() method.
	context math.Matrix
}

func (f *TriangleSoup) NumTriangles() uint64 {
	return uint64(len(f.faces))
}

func (f *TriangleSoup) Faces(iter func(primitive.Face, material.Material) bool) {
	for i := range f.faces {
		if !iter(f.faces[i], f.material) {
			return
		}
	}
}

func (f *TriangleSoup) GetMaterial() material.Material {
	return f.material
}

// NewTriangleSoup returns a triangular soup.
func NewTriangleSoup(ts []*primitive.Triangle) *TriangleSoup {
	// Compute AABB at loading time.
	aabb := ts[0].AABB()
	for i := 1; i < len(ts); i++ {
		aabb.Add(ts[i].AABB())
	}

	return &TriangleSoup{
		faces:   ts,
		aabb:    aabb,
		context: math.MatI,
	}
}

func (t *TriangleSoup) UseMaterial(mat material.Material) {
	t.material = mat
}

// modelMatrix returns the transformation context as the model matrix
// for the current frame (or at call time).
func (t *TriangleSoup) ModelMatrix() math.Matrix {
	return t.context
}

func (t *TriangleSoup) ResetContext() {
	t.context = math.MatI
}

// Scale sets the scale matrix.
func (m *TriangleSoup) Scale(sx, sy, sz float64) {
	m.context = math.NewMatrix(
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(m.context)
}

// SetTranslate sets the translate matrix.
func (m *TriangleSoup) Translate(tx, ty, tz float64) {
	m.context = math.NewMatrix(
		1, 0, 0, tx,
		0, 1, 0, ty,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(m.context)
}

func (m *TriangleSoup) Rotate(dir math.Vector, angle float64) {
	u := dir.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	m.context = q.ToRoMat().MulM(m.context)
}

func (m *TriangleSoup) RotateX(angle float64) {
	u := math.NewVector(1, 0, 0, 0)
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	m.context = q.ToRoMat().MulM(m.context)
}

func (m *TriangleSoup) RotateY(angle float64) {
	u := math.NewVector(0, 1, 0, 0)
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	m.context = q.ToRoMat().MulM(m.context)
}

func (m *TriangleSoup) RotateZ(angle float64) {
	u := math.NewVector(0, 0, 1, 0)
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	m.context = q.ToRoMat().MulM(m.context)
}

func (m *TriangleSoup) AABB() primitive.AABB {
	min := m.aabb.Min.Apply(m.ModelMatrix())
	max := m.aabb.Max.Apply(m.ModelMatrix())
	return primitive.AABB{Min: min, Max: max}
}

func (m *TriangleSoup) Center() math.Vector {
	aabb := m.AABB()
	return aabb.Min.Add(aabb.Max).Pos()
}

func (m *TriangleSoup) Radius() float64 {
	aabb := m.AABB()
	return aabb.Max.Sub(aabb.Min).Len() / 2
}

// Normalize rescales the mesh to the unit sphere centered at the origin.
func (m *TriangleSoup) Normalize() {
	aabb := m.AABB()
	center := aabb.Min.Add(aabb.Max).Pos()
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	for i := 0; i < len(m.faces); i++ {
		f := m.faces[i]
		f.V1.Pos = f.V1.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		f.V2.Pos = f.V2.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
		f.V3.Pos = f.V3.Pos.Apply(m.ModelMatrix()).Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	}

	// update AABB after scaling
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac, 1)
	m.aabb = primitive.AABB{Min: min, Max: max}
	m.ResetContext()
}
