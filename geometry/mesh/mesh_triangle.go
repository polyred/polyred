// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

var _ Mesh = &TriangleMesh{}

// TriangleMesh implements a triangular mesh.
type TriangleMesh struct {
	ibo buffer.IndexBuffer
	vbo buffer.VertexBuffer

	// caches
	tris []*primitive.Triangle
	aabb *primitive.AABB
}

func (m *TriangleMesh) Triangles() []*primitive.Triangle { return m.tris }

// NewTriangleMesh returns a triangular soup.
func NewTriangleMesh(tris []*primitive.Triangle) *TriangleMesh {
	if len(tris) == 0 {
		panic("mesh: cannot construct a triangle mesh without any faces")
	}

	ibo := make([]int, len(tris)*3)
	for i := 0; i < len(ibo); i++ {
		ibo[i] = i
	}
	vbo := make([]*primitive.Vertex, len(tris)*3)
	for i := 0; i < len(tris); i++ {
		vbo[3*i+0] = tris[i].V1
		vbo[3*i+1] = tris[i].V2
		vbo[3*i+2] = tris[i].V3
	}

	// Compute AABB at loading time.
	aabb := tris[0].AABB()
	for i := 1; i < len(tris); i++ {
		aabb.Add(tris[i].AABB())
	}

	return &TriangleMesh{
		ibo: ibo,
		vbo: vbo,

		tris: tris,
		aabb: &aabb,
	}
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

	return primitive.AABB{Min: m.aabb.Min, Max: m.aabb.Max}
}

func (m *TriangleMesh) Center() math.Vec3[float32] {
	aabb := m.AABB()
	return aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
}

func (m *TriangleMesh) Radius() float32 {
	aabb := m.AABB()
	return aabb.Max.Sub(aabb.Min).Len() / 2
}
