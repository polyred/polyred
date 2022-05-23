// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package object

import (
	"poly.red/geometry/primitive"
	"poly.red/math"
)

type Type int

const (
	TypeGroup = iota
	TypeGeometry
	TypeCamera
	TypeLight
)

type Object[T math.Float] interface {
	Name() string

	Type() Type
	Rotate(dir math.Vec3[T], angle T)
	RotateX(a T)
	RotateY(a T)
	RotateZ(a T)
	Translate(x, y, z T)
	Scale(x, y, z T)
	AABB() primitive.AABB
	ModelMatrix() math.Mat4[T]
}
