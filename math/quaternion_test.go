// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"testing"

	"changkun.de/x/polyred/math"
)

func TestQuaternionToRotationMatrix(t *testing.T) {
	dirX := math.Vec4{1, 0, 0, 0}
	angle := math.Pi / 3

	u := dirX.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)

	want := math.Mat4{
		1, 0, 0, 0,
		0, 0.5, -0.8660254, 0,
		0, 0.8660254, 0.5, 0,
		0, 0, 0, 1,
	}
	got := q.ToRoMat()
	if !got.Eq(want) {
		t.Fatalf("ToRoMat is wrong, want: %v, got: %v", want, got)
	}

	dirY := math.Vec4{0, 1, 0, 0}
	u = dirY.Unit()
	cosa = math.Cos(angle / 2)
	sina = math.Sin(angle / 2)
	q = math.Quaternion{cosa, math.Vec4{sina * u.X, sina * u.Y, sina * u.Z, 0}}
	want = math.Mat4{
		0.5, 0, 0.8660254, 0,
		0, 1, 0, 0,
		-0.8660254, 0, 0.5, 0,
		0, 0, 0, 1,
	}
	got = q.ToRoMat()
	if !got.Eq(want) {
		t.Fatalf("ToRoMat is wrong, want: %v, got: %v", want, got)
	}

	dirZ := math.Vec4{0, 0, 1, 0}
	u = dirZ.Unit()
	cosa = math.Cos(angle / 2)
	sina = math.Sin(angle / 2)
	q = math.Quaternion{cosa, math.Vec4{sina * u.X, sina * u.Y, sina * u.Z, 0}}
	want = math.Mat4{
		0.5, -0.8660254, 0, 0,
		0.8660254, 0.5, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	got = q.ToRoMat()
	if !got.Eq(want) {
		t.Fatalf("ToRoMat is wrong, want: %v, got: %v", want, got)
	}
}

func BenchmarkQuaternion_ToRoMat(b *testing.B) {
	dirX := math.Vec4{1, 0, 0, 0}
	angle := math.Pi / 3

	u := dirX.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = q.ToRoMat()
	}
	_ = m
}
