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
	Rotate(dir math.Vec3, angle float32)
	RotateX(a float32)
	RotateY(a float32)
	RotateZ(a float32)
	Translate(x, y, z float32)
	Scale(x, y, z float32)
	ModelMatrix() math.Mat4
}
