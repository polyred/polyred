// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
)

var _ Mesh = &PolygonMesh{}

// PolygonMesh is a polygon based mesh structure that can contain
// arbitrary shaped faces, such as triangle and quad mixed mesh.
type PolygonMesh struct {
	ibo buffer.IndexBuffer
	vbo buffer.VertexBuffer

	// caches
	polys []*primitive.Polygon
	aabb  *primitive.AABB
}

func NewPolygonMesh(faces []primitive.Face) *PolygonMesh {
	if len(faces) == 0 {
		panic("mesh: cannot construct a quad mesh without any faces")
	}

	ibo := []int{}
	vbo := []*primitive.Vertex{}
	polys := []*primitive.Polygon{}
	i := 0
	for _, f := range faces {
		vs := []*primitive.Vertex{}
		f.Vertices(func(v *primitive.Vertex) bool {
			ibo = append(ibo, i)
			vbo = append(vbo, v)
			vs = append(vs, v)
			return true
		})
		polys = append(polys, primitive.NewPolygon(vs...))
	}

	// Compute AABB at loading time.
	aabb := faces[0].AABB()
	for i := 1; i < len(faces); i++ {
		aabb.Add(faces[i].AABB())
	}

	return &PolygonMesh{
		ibo: ibo,
		vbo: vbo,

		polys: polys,
		aabb:  &aabb,
	}
}

func (m *PolygonMesh) AABB() primitive.AABB {
	if m.aabb == nil {
		// Compute AABB if not computed
		aabb := m.polys[0].AABB()
		lenth := len(m.polys)
		for i := 1; i < lenth; i++ {
			aabb.Add(m.polys[i].AABB())
		}
		m.aabb = &aabb
	}

	return primitive.AABB{Min: m.aabb.Min, Max: m.aabb.Max}
}

func (m *PolygonMesh) Triangles() []*primitive.Triangle {
	tris := []*primitive.Triangle{}
	for _, poly := range m.polys {
		poly.Triangles(func(t *primitive.Triangle) bool {
			tris = append(tris, t)
			return true
		})
	}
	return tris
}
