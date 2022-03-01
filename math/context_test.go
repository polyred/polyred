// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"testing"

	"poly.red/math"
)

func TestTransformationContext(t *testing.T) {

	ctx := math.TransformContext{}
	ctx.ResetContext()

	ctx.Scale(1, 2, 3)
	ctx.Translate(1, 2, 3)

	modMat := ctx.ModelMatrix()
	want := math.NewMat4(
		1, 0, 0, 1,
		0, 2, 0, 2,
		0, 0, 3, 3,
		0, 0, 0, 1,
	)
	if !modMat.Eq(want) {
		t.Fatalf("unexpected model matrix, got %v, want %v", modMat, want)
	}

	ctx.Scale(1, 2, 3)
	modMat = ctx.ModelMatrix()
	want = math.NewMat4(
		1, 0, 0, 1,
		0, 4, 0, 2,
		0, 0, 9, 3,
		0, 0, 0, 1,
	)
	if !modMat.Eq(want) {
		t.Fatalf("unexpected model matrix, got %v, want %v", modMat, want)
	}

	ctx.ResetContext()
	ctx.Rotate(math.NewVec3(0, 1, 0), math.HalfPi)
	modMat = ctx.ModelMatrix()
	want = math.NewMat4(
		0, 0, 1, 0,
		0, 1, 0, 0,
		-1, 0, 0, 0,
		0, 0, 0, 1,
	)
	if !modMat.Eq(want) {
		t.Fatalf("unexpected model matrix, got %v, want %v", modMat, want)
	}

	ctx.ResetContext()
	ctx.Rotate(math.NewVec3(1, 0, 0), math.HalfPi)
	modMat = ctx.ModelMatrix()
	want = math.NewMat4(
		1, 0, 0, 0,
		0, 0, -1, 0,
		0, 1, 0, 0,
		0, 0, 0, 1,
	)
	if !modMat.Eq(want) {
		t.Fatalf("unexpected model matrix, got %v, want %v", modMat, want)
	}

	ctx.ResetContext()
	ctx.Rotate(math.NewVec3(0, 0, 1), math.HalfPi)
	modMat = ctx.ModelMatrix()
	want = math.NewMat4(
		0, -1, 0, 0,
		1, 0, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	)
	if !modMat.Eq(want) {
		t.Fatalf("unexpected model matrix, got %v, want %v", modMat, want)
	}
}

func TestTransformationContextScale(t *testing.T) {

	t.Run("scale", func(t *testing.T) {

		ctx1 := math.TransformContext{}
		ctx1.ResetContext()
		ctx2 := math.TransformContext{}
		ctx2.ResetContext()

		ctx1.Scale(2, 3, 4)
		want := ctx1.ModelMatrix()

		ctx2.ScaleX(2)
		ctx2.ScaleY(3)
		ctx2.ScaleZ(4)

		got := ctx2.ModelMatrix()
		if !want.Eq(got) {
			t.Fatalf("unexpected model matrix, got %v, want %v", got, want)
		}

	})

	t.Run("translate", func(t *testing.T) {

		ctx1 := math.TransformContext{}
		ctx1.ResetContext()
		ctx2 := math.TransformContext{}
		ctx2.ResetContext()

		ctx1.Translate(2, 3, 4)
		want := ctx1.ModelMatrix()

		ctx2.TranslateX(2)
		ctx2.TranslateY(3)
		ctx2.TranslateZ(4)

		got := ctx2.ModelMatrix()
		if !want.Eq(got) {
			t.Fatalf("unexpected model matrix, got %v, want %v", got, want)
		}

	})

	t.Run("rotateX", func(t *testing.T) {

		ctx1 := math.TransformContext{}
		ctx1.ResetContext()
		ctx2 := math.TransformContext{}
		ctx2.ResetContext()

		ctx1.Rotate(math.NewVec3(1, 0, 0), math.HalfPi)
		want := ctx1.ModelMatrix()

		ctx2.RotateX(math.HalfPi)

		got := ctx2.ModelMatrix()
		if !want.Eq(got) {
			t.Fatalf("unexpected model matrix, got %v, want %v", got, want)
		}

	})

	t.Run("rotateY", func(t *testing.T) {

		ctx1 := math.TransformContext{}
		ctx1.ResetContext()
		ctx2 := math.TransformContext{}
		ctx2.ResetContext()

		ctx1.Rotate(math.NewVec3(0, 1, 0), math.HalfPi)
		want := ctx1.ModelMatrix()

		ctx2.RotateY(math.HalfPi)

		got := ctx2.ModelMatrix()
		if !want.Eq(got) {
			t.Fatalf("unexpected model matrix, got %v, want %v", got, want)
		}

	})

	t.Run("rotateZ", func(t *testing.T) {

		ctx1 := math.TransformContext{}
		ctx1.ResetContext()
		ctx2 := math.TransformContext{}
		ctx2.ResetContext()

		ctx1.Rotate(math.NewVec3(0, 0, 1), math.HalfPi)
		want := ctx1.ModelMatrix()

		ctx2.RotateZ(math.HalfPi)

		got := ctx2.ModelMatrix()
		if !want.Eq(got) {
			t.Fatalf("unexpected model matrix, got %v, want %v", got, want)
		}
	})
}
