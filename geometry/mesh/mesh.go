// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package mesh represents polygon based mesh objects.
//
// Note that a mesh object cannot be transformed unless it is turned
// to a geometry.Geometry object. See package geometry for more info.
package mesh

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

type Mesh[T math.Float] interface {
	AABB() primitive.AABB

	IndexBuffer() buffer.IndexBuffer
	VertexBuffer() buffer.VertexBuffer
	Triangles() []*primitive.Triangle
}
