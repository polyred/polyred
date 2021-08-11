// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	_ "image/jpeg" // for jpg encoding

	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/object"
)

type Mesh interface {
	object.Object

	AABB() primitive.AABB
	Normalize()
	SetMaterial(m material.Material)
	GetMaterial() material.Material
	NumTriangles() uint64
	Faces(func(f primitive.Face, m material.Material) bool)
}
