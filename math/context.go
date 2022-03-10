// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
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
type TransformContext[T Float] struct {
	context    Mat4[T]
	needUpdate bool

	// We use a quaternion to persist the rotation context, so that we
	// don't have the Gimbal Lock issue.
	//
	// See https://en.wikipedia.org/wiki/Gimbal_lock.
	rotation  Quaternion[T]
	scale     Mat4[T]
	translate Mat4[T]
}

// ModelMatrix returns the most recent transformation context.
func (ctx *TransformContext[T]) ModelMatrix() Mat4[T] {
	if ctx.needUpdate {
		ctx.context = ctx.translate.MulM(ctx.scale).MulM(ctx.rotation.ToRoMat())
		ctx.needUpdate = false
	}
	return ctx.context
}

// ResetContext resets the transformation context.
func (ctx *TransformContext[T]) ResetContext() {
	ctx.context = Mat4I[T]()
	ctx.rotation = NewQuaternion[T](1, 0, 0, 0)
	ctx.scale = Mat4I[T]()
	ctx.translate = Mat4I[T]()
	ctx.needUpdate = false
}

// Scale sets the scale matrix.
func (ctx *TransformContext[T]) Scale(sx, sy, sz T) {
	ctx.scale = NewMat4(
		sx, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleX sets the scale matrix on X-axis.
func (ctx *TransformContext[T]) ScaleX(sx T) {
	ctx.scale = NewMat4(
		sx, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleY sets the scale matrix on Y-axis.
func (ctx *TransformContext[T]) ScaleY(sy T) {
	ctx.scale = NewMat4(
		1, 0, 0, 0,
		0, sy, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// ScaleZ sets the scale matrix on Z-axis.
func (ctx *TransformContext[T]) ScaleZ(sz T) {
	ctx.scale = NewMat4(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, sz, 0,
		0, 0, 0, 1,
	).MulM(ctx.scale)
	ctx.needUpdate = true
}

// Translate sets the translate matrix.
func (ctx *TransformContext[T]) Translate(tx, ty, tz T) {
	ctx.translate = NewMat4(
		1, 0, 0, tx,
		0, 1, 0, ty,
		0, 0, 1, tz,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateX sets the translate matrix on X-axis.
func (ctx *TransformContext[T]) TranslateX(tx T) {
	ctx.translate = NewMat4(
		1, 0, 0, tx,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateY sets the translate matrix on Y-axis.
func (ctx *TransformContext[T]) TranslateY(ty T) {
	ctx.translate = NewMat4(
		1, 0, 0, 0,
		0, 1, 0, ty,
		0, 0, 1, 0,
		0, 0, 0, 1,
	).MulM(ctx.translate)
	ctx.needUpdate = true
}

// TranslateZ sets the translate matrix on Z-axis.
func (ctx *TransformContext[T]) TranslateZ(tz T) {
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
func (ctx *TransformContext[T]) Rotate(dir Vec3[T], angle T) {
	u := dir.Unit()
	cosa := Cos(angle * 0.5)
	sina := Sin(angle * 0.5)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateX applies rotation on X-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext[T]) RotateX(angle T) {
	u := NewVec3[T](1, 0, 0)
	cosa := Cos(angle * 0.5)
	sina := Sin(angle * 0.5)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateY applies rotation on Y-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext[T]) RotateY(angle T) {
	u := NewVec3[T](0, 1, 0)
	cosa := Cos(angle * 0.5)
	sina := Sin(angle * 0.5)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}

// RotateZ applies rotation on Z-axis direction with an specified
// angle counterclockwise.
func (ctx *TransformContext[T]) RotateZ(angle T) {
	u := NewVec3[T](0, 0, 1)
	cosa := Cos(angle * 0.5)
	sina := Sin(angle * 0.5)
	q := NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)
	ctx.rotation = q.Mul(ctx.rotation)
	ctx.needUpdate = true
}
