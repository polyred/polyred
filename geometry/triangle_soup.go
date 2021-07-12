// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

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
	aabb *primitive.AABB

	math.TransformContext
}

func (f *TriangleSoup) Type() object.Type {
	return object.TypeMesh
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

func (t *TriangleSoup) SetMaterial(mat material.Material) {
	t.material = mat
}

// NewTriangleSoup returns a triangular soup.
func NewTriangleSoup(ts []*primitive.Triangle) *TriangleSoup {
	// Compute AABB at loading time.
	aabb := ts[0].AABB()
	for i := 1; i < len(ts); i++ {
		aabb.Add(ts[i].AABB())
	}

	ret := &TriangleSoup{
		faces: ts,
		aabb:  &aabb,
	}
	ret.ResetContext()
	return ret
}

func (m *TriangleSoup) AABB() primitive.AABB {
	if m.aabb == nil {
		// Compute AABB if not computed
		aabb := m.faces[0].AABB()
		lenth := len(m.faces)
		for i := 1; i < lenth; i++ {
			aabb.Add(m.faces[i].AABB())
		}
		m.aabb = &aabb
	}

	min := m.aabb.Min.ToVec4(1).Apply(m.ModelMatrix()).ToVec3()
	max := m.aabb.Max.ToVec4(1).Apply(m.ModelMatrix()).ToVec3()
	return primitive.AABB{Min: min, Max: max}
}

func (m *TriangleSoup) Center() math.Vec3 {
	aabb := m.AABB()
	return aabb.Min.Add(aabb.Max).Scale(1/2, 1/2, 1/2)
}

func (m *TriangleSoup) Radius() float64 {
	aabb := m.AABB()
	return aabb.Max.Sub(aabb.Min).Len() / 2
}

// Normalize rescales the mesh to the unit sphere centered at the origin.
func (m *TriangleSoup) Normalize() {
	aabb := m.AABB()
	center := aabb.Min.Add(aabb.Max).Scale(1/2, 1/2, 1/2)
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
	min := aabb.Min.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
	max := aabb.Max.Translate(-center.X, -center.Y, -center.Z).Scale(fac, fac, fac)
	m.aabb = &primitive.AABB{Min: min, Max: max}
	m.ResetContext()
}

func (m *TriangleSoup) GetVertexIndex() []uint64 {
	index := make([]uint64, len(m.faces)*3)

	for i := range index {
		index[i] = uint64(i)
	}
	return index
}

func (m *TriangleSoup) GetVertexBuffer() []*primitive.Vertex {
	vs := make([]*primitive.Vertex, len(m.faces)*3)
	i := 0
	m.Faces(func(f primitive.Face, m material.Material) bool {
		f.Vertices(func(v *primitive.Vertex) bool {
			vs[i] = v
			i++
			return true
		})
		return true
	})
	return vs
}
