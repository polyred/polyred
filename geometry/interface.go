package geometry

import "poly.red/geometry/primitive"

type Renderable interface {
	PrimitiveType() primitive.Type
	IndexBuffer() []uint64
	VertexBuffer() []*primitive.Vertex
}
