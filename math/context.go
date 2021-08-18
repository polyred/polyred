// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

// TransformContext is a transformation context (model matrix) that
// accumulates applied transformation matrices. The applying order of
// different types of transformations is:
//
// 1. rotation, 2. scaling, and 3. translation
//
// context is a persistent status for the given mesh and can be reused
// for each of the rendering frame unless the mesh intentionally calls
// ResetContext() method.
//
// A transformation context must be reset before use.
type TransformContext struct {
	context    Mat4
	needUpdate bool

	// We use a quaternion to persist the rotation context, so that we
	// don't have the Gimbal Lock issue.
	//
	// See https://en.wikipedia.org/wiki/Gimbal_lock.
	rotation  Quaternion
	scale     Mat4
	translate Mat4
}

// ModelMatrix returns the most recent transformation context.
func (ctx *TransformContext) ModelMatrix() Mat4 {
	if ctx.needUpdate {
		ctx.context = ctx.translate.MulM(ctx.scale).MulM(ctx.rotation.ToRoMat())
		ctx.needUpdate = false
	}
	return ctx.context
}

// ResetContext resets the transformation context.
func (ctx *TransformContext) ResetContext() {
	ctx.context = Mat4I
	ctx.rotation = NewQuaternion(1, 0, 0, 0)
	ctx.scale = Mat4I
	ctx.translate = Mat4I
	ctx.needUpdate = false
}

// Scale sets the scale matrix.
func (ctx *TransformContext) Scale(sx, sy, sz float64) {
	ctx.scale = NewMat4(
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleX sets the scale matrix on X-axis.
func (ctx *TransformContext) ScaleX(sx float64) {
	ctx.scale = NewMat4(
		sx, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleY sets the scale matrix on Y-axis.
func (ctx *TransformContext) ScaleY(sy float64) {
	ctx.scale = NewMat4(
		1, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleZ sets the scale matrix on Z-axis.
func (ctx *TransformContext) ScaleZ(sz float64) {
	ctx.scale = NewMat4(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// Translate sets the translate matrix.
func (ctx *TransformContext) Translate(tx, ty, tz float64) {
	ctx.translate = NewMat4(
		1, 0, 0, tx,
		0, 1, 0, ty,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateX sets the translate matrix on X-axis.
func (ctx *TransformContext) TranslateX(tx float64) {
	ctx.translate = NewMat4(
		1, 0, 0, tx,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateY sets the translate matrix on Y-axis.
func (ctx *TransformContext) TranslateY(ty float64) {
	ctx.translate = NewMat4(
		1, 0, 0, 0,
		0, 1, 0, ty,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateZ sets the translate matrix on Z-axis.
func (ctx *TransformContext) TranslateZ(tz float64) {
	ctx.translate = NewMat4(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// Rotate applies rotation on an arbitrary direction with an specified
// angle counterclockwise.
func (ctx *TransformContext) Rotate(dir Vec3, angle float64) {
	u := dir.Unit()
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateX applies rotation on X-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext) RotateX(angle float64) {
	u := NewVec3(1, 0, 0)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateY applies rotation on Y-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext) RotateY(angle float64) {
	u := NewVec3(0, 1, 0)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateZ applies rotation on Z-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext) RotateZ(angle float64) {
	u := NewVec3(0, 0, 1)
	cosa := Cos(angle / 2)
	sina := Sin(angle / 2)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}
