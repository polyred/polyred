package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
)

var _ Mesh[float32] = &QuadMesh{}

type QuadMesh struct {
}

func NewQuadMesh(quads []*primitive.Quad) *QuadMesh {
	panic("unimplemented")
}

func (m *QuadMesh) AABB() primitive.AABB              { panic("unimplemented") }
func (m *QuadMesh) Normalize()                        { panic("unimplemented") }
func (m *QuadMesh) IndexBuffer() buffer.IndexBuffer   { panic("unimplemented") }
func (m *QuadMesh) VertexBuffer() buffer.VertexBuffer { panic("unimplemented") }
func (m *QuadMesh) Triangles() []*primitive.Triangle  { panic("unimplemented") }
