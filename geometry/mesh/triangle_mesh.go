// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ Mesh[float32] = &TriangleMesh{}
)

// TriangleMesh implements a triangular mesh.
type TriangleMesh struct { // TriangleSoup
	ibo buffer.IndexBuffer
	vbo buffer.VertexBuffer
	mat material.BlinnPhong

	// caches
	tris []*primitive.Triangle
	aabb *primitive.AABB

	math.TransformContext[float32]
}

func (f *TriangleMesh) Name() string                      { return "triangle_mesh" }
func (f *TriangleMesh) Type() object.Type                 { return object.TypeMesh }
func (f *TriangleMesh) Triangles() []*primitive.Triangle  { return f.tris }
func (m *TriangleMesh) IndexBuffer() buffer.IndexBuffer   { return m.ibo }
func (m *TriangleMesh) VertexBuffer() buffer.VertexBuffer { return m.vbo }

// NewTriangleMesh returns a triangular soup.
func NewTriangleMesh(ts []*primitive.Triangle) *TriangleMesh {
	ibo := make([]int, len(ts)*3)
	for i := 0; i < len(ibo); i++ {
		ibo[i] = i
	}
	vbo := make([]*primitive.Vertex, len(ts)*3)
	for i := 0; i < len(ts); i++ {
		vbo[3*i+0] = ts[i].V1
		vbo[3*i+1] = ts[i].V2
		vbo[3*i+2] = ts[i].V3
	}

	// Compute AABB at loading time.
	aabb := ts[0].AABB()
	for i := 1; i < len(ts); i++ {
		aabb.Add(ts[i].AABB())
	}

	ret := &TriangleMesh{
		ibo: ibo,
		vbo: vbo,

		tris: ts,
		aabb: &aabb,
	}
	ret.ResetContext()
	return ret
}

func (m *TriangleMesh) AABB() primitive.AABB {
	if m.aabb == nil {
		// Compute AABB if not computed
		aabb := m.tris[0].AABB()
		lenth := len(m.tris)
		for i := 1; i < lenth; i++ {
			aabb.Add(m.tris[i].AABB())
		}
		m.aabb = &aabb
	}

	min := m.aabb.Min.ToVec4(1).Apply(m.ModelMatrix()).ToVec3()
	max := m.aabb.Max.ToVec4(1).Apply(m.ModelMatrix()).ToVec3()
	return primitive.AABB{Min: min, Max: max}
}

func (m *TriangleMesh) Center() math.Vec3[float32] {
	aabb := m.AABB()
	return aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
}

func (m *TriangleMesh) Radius() float32 {
	aabb := m.AABB()
	return aabb.Max.Sub(aabb.Min).Len() / 2
}

// Normalize rescales the mesh to the unit sphere centered at the origin.
func (m *TriangleMesh) Normalize() {
	aabb := m.AABB()
	center := aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	fac := 1 / radius

	// scale all vertices
	for i := 0; i < len(m.tris); i++ {
		f := m.tris[i]
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
