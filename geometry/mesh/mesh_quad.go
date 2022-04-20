package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
)

var _ Mesh = &QuadMesh{}

type QuadMesh struct {
	ibo buffer.IndexBuffer
	vbo buffer.VertexBuffer

	// caches
	quads []*primitive.Quad
	aabb  *primitive.AABB
}

func NewQuadMesh(quads []*primitive.Quad) *QuadMesh {
	if len(quads) == 0 {
		panic("mesh: cannot construct a quad mesh without any faces")
	}

	ibo := make([]int, len(quads)*4)
	for i := 0; i < len(ibo); i++ {
		ibo[i] = i
	}
	vbo := make([]*primitive.Vertex, len(quads)*4)
	for i := 0; i < len(quads); i++ {
		vbo[4*i+0] = quads[i].V1
		vbo[4*i+1] = quads[i].V2
		vbo[4*i+2] = quads[i].V3
		vbo[4*i+3] = quads[i].V4
	}

	// Compute AABB at loading time.
	aabb := quads[0].AABB()
	for i := 1; i < len(quads); i++ {
		aabb.Add(quads[i].AABB())
	}

	return &QuadMesh{
		ibo: ibo,
		vbo: vbo,

		quads: quads,
		aabb:  &aabb,
	}
}

func (m *QuadMesh) Triangles() []*primitive.Triangle {
	tris := []*primitive.Triangle{}
	for _, quad := range m.quads {
		quad.Triangles(func(t *primitive.Triangle) bool {
			tris = append(tris, t)
			return true
		})
	}
	return tris
}

func (m *QuadMesh) AABB() primitive.AABB {
	if m.aabb == nil {
		// Compute AABB if not computed
		aabb := m.quads[0].AABB()
		lenth := len(m.quads)
		for i := 1; i < lenth; i++ {
			aabb.Add(m.quads[i].AABB())
		}
		m.aabb = &aabb
	}

	return primitive.AABB{Min: m.aabb.Min, Max: m.aabb.Max}

}
