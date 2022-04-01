// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	_ "image/jpeg" // for jpg encoding

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
	"poly.red/scene/object"
)

type Mesh[T math.Float] interface {
	object.Object[T]

	AABB() primitive.AABB
	Normalize()
	SetMaterial(m material.Material)
	GetMaterial() material.Material

	IndexBuffer() buffer.IndexBuffer
	VertexBuffer() buffer.VertexBuffer
	Triangles() []*primitive.Triangle
}

type MeshPointer[T Mesh[float32]] interface {
	*T
}
