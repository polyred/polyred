// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package object

import "poly.red/math"

type Type int

const (
	TypeGroup = iota
	TypeMesh
	TypeCamera
	TypeLight
)

type Object interface {
	Type() Type
	Rotate(dir math.Vec3, angle float64)
	RotateX(a float64)
	RotateY(a float64)
	RotateZ(a float64)
	Translate(x, y, z float64)
	Scale(x, y, z float64)
	ModelMatrix() math.Mat4
}
