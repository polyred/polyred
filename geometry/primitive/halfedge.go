// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

var _ Edge[float32] = &Halfedge{}

// Halfedge holds a halfedge representation.
type Halfedge struct {
	Idx  uint64
	Next *Halfedge
	Prev *Halfedge
	Twin *Halfedge
	Face Face[float32]

	v          *Vertex
	onBoundary bool
}

// Verts returns the touching vertices of the given halfedge.
func (he *Halfedge) Verts() (v1, v2 *Vertex) {
	return he.v, he.Next.v
}

func (he *Halfedge) Vec() math.Vec4[float32] {
	return he.Next.v.Pos.Sub(he.v.Pos)
}

func (he *Halfedge) DihedralAngle() float32 {
	if he.onBoundary || he.Twin.onBoundary {
		return 0
	}

	if he.Face != nil && he.Twin != nil && he.Twin.Face != nil {
		n1 := he.Face.Normal()
		n2 := he.Twin.Face.Normal()
		w := he.Vec().Unit()
		return math.Atan2(n1.Cross(n2).Dot(w), n1.Dot(n2))
	}

	return 0
}

func (he *Halfedge) Cotan() float32 {
	if he.onBoundary {
		return 0
	}

	u := he.Prev.Vec()
	v := he.Next.Vec().Scale(-1, -1, -1, 1)
	return u.Dot(v) / u.Cross(v).Len()
}

func (he *Halfedge) OnBoundary() bool {
	return he.onBoundary
}
