// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package camera provides a camera abstraction for perspective
// and orthographic camera and their utilities, such as viewing
// transformation matrices.
package camera

import (
	"poly.red/math"
	"poly.red/object"
)

// Interface assertion
var (
	_ Interface = &Orthographic{}
	_ Interface = &Perspective{}
)

// Interface is a camera interface that represents either orthographic
// or perspective camera.
type Interface interface {
	object.Object
	Type() object.Type

	Fov() float32
	Aspect() float32
	SetAspect(float32, float32)
	Position() math.Vec3
	SetPosition(math.Vec3)
	LookAt() (math.Vec3, math.Vec3)
	SetLookAt(math.Vec3, math.Vec3)
	ViewMatrix() math.Mat4
	ProjMatrix() math.Mat4
}

// ViewMatrix is a handy function for computing view matrix without
// instantiating all required camera parameters. The camera view matrix
// is determined via its position, look at target position, and the up
// direction.
func ViewMatrix(pos, target, up math.Vec3) math.Mat4 {
	l := target.Sub(pos).Unit()
	lxu := l.Cross(up).Unit()
	u := lxu.Cross(l).Unit()
	x := pos.X
	y := pos.Y
	z := pos.Z
	TrTt := math.NewMat4(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt
}
