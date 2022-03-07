package primitive

import "poly.red/math"

// Edge is an interface that abstracts any edge representations.
type Edge interface {
	Verts() (*Vertex, *Vertex)
}

// Face is a polygon face that abstracts any face representations.
type Face interface {
	Normal() math.Vec4
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}
