// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

// TransformContext is a transformation context (model matrix) that
// accumulates applied transformation matrices (multiplied from left side).
//
// context is a persistant status for the given mesh and can be reused
// for each of the rendering frame unless the mesh intentionally calls
// ResetContext() method.
type TransformContext struct {
	context Matrix
}

func (ctx *TransformContext) ModelMatrix() Matrix {
	return ctx.context
}

func (ctx *TransformContext) ResetContext() {
	ctx.context = MatI
}

// Scale sets the scale matrix.
func (ctx *TransformContext) Scale(sx, sy, sz float64) {
	ctx.context = NewMatrix(
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(ctx.context)
}

// SetTranslate sets the translate matrix.
func (ctx *TransformContext) Translate(tx, ty, tz float64) {
	ctx.context = NewMatrix(
		1, 0, 0, tx,
		0, 1, 0, ty,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(ctx.context)
}

func (ctx *TransformContext) Rotate(dir Vector, angle float64) {
	u := dir.Unit()
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.context = q.ToRoMat().MulM(ctx.context)
}

func (ctx *TransformContext) RotateX(angle float64) {
	u := NewVector(1, 0, 0, 0)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.context = q.ToRoMat().MulM(ctx.context)
}

func (ctx *TransformContext) RotateY(angle float64) {
	u := NewVector(0, 1, 0, 0)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.context = q.ToRoMat().MulM(ctx.context)
}

func (ctx *TransformContext) RotateZ(angle float64) {
	u := NewVector(0, 0, 1, 0)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.context = q.ToRoMat().MulM(ctx.context)
}
