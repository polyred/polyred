// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

// Halfedge holds a halfedge representation.
type Halfedge struct {
	v          *Vertex
	f          Face
	prev       *Halfedge
	next       *Halfedge
	twin       *Halfedge
	idx        int64
	onBoundary bool
}

func (he *Halfedge) Vec() math.Vec4 {
	return he.next.v.Pos.Sub(he.v.Pos)
}

func (he *Halfedge) DihedralAngle() float32 {
	if he.onBoundary || he.twin.onBoundary {
		return 0
	}

	n1 := he.f.Normal()
	n2 := he.twin.f.Normal()
	w := he.Vec().Unit()
	return math.Atan2(n1.Cross(n2).Dot(w), n1.Dot(n2))
}

func (he *Halfedge) Cotan() float32 {
	if he.onBoundary {
		return 0
	}

	u := he.prev.Vec()
	v := he.next.Vec().Scale(-1, -1, -1, 1)
	return u.Dot(v) / u.Cross(v).Len()
}
