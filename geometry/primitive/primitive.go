// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

// Edge is an interface that abstracts any edge representations.
type Edge interface {
	Verts() (*Vertex, *Vertex)
}

// Face is a polygon face that abstracts any face representations.
type Face interface {
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}
