package primitive

import "poly.red/math"

// Edge is an interface that abstracts any edge representations.
type Edge[T math.Float] interface {
	Verts() (*Vertex, *Vertex)
}

// Face is a polygon face that abstracts any face representations.
type Face[T math.Float] interface {
	Normal() math.Vec4[float32]
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}
