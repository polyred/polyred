// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package camera provides a camera abstraction for perspective
// and orthographic camera and their utilities, such as viewing
// transformation matrices.
package camera

import (
	"poly.red/math"
	"poly.red/scene/object"
)

// Interface assertion
var (
	_ Interface = &Orthographic{}
	_ Interface = &Perspective{}
)

// Interface is a camera interface that represents either orthographic
// or perspective camera.
type Interface interface {
	object.Object[float32]
	Type() object.Type

	Fov() float32
	Aspect() float32
	SetAspect(float32, float32)
	Position() math.Vec3[float32]
	SetPosition(math.Vec3[float32])
	LookAt() (math.Vec3[float32], math.Vec3[float32])
	SetLookAt(math.Vec3[float32], math.Vec3[float32])
	ViewMatrix() math.Mat4[float32]
	ProjMatrix() math.Mat4[float32]
}

// ViewMatrix is a handy function for computing view matrix without
// instantiating all required camera parameters. The camera view matrix
// is determined via its position, look at target position, and the up
// direction.
func ViewMatrix(pos, target, up math.Vec3[float32]) math.Mat4[float32] {
	l := target.Sub(pos).Unit()
	lxu := l.Cross(up).Unit()
	u := lxu.Cross(l).Unit()
	TrTt := math.NewMat4(
		lxu.X, lxu.Y, lxu.Z, -lxu.Dot(pos),
		u.X, u.Y, u.Z, -u.Dot(pos),
		-l.X, -l.Y, -l.Z, l.Dot(pos),
		0, 0, 0, 1,
	)
	return TrTt
}
