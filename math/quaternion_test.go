// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package math_test

import (
	"testing"

	"changkun.de/x/ddd/math"
)

func TestQuaternionToRotationMatrix(t *testing.T) {
	dirX := math.Vector{1, 0, 0, 0}
	angle := math.Pi / 3

	u := dirX.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)

	want := math.Matrix{
		1,
		0,
		0,
		0,
		0,
		0.5,
		-0.8660254,
		0,
		0,
		0.8660254,
		0.5,
		0,
		0,
		0,
		0,
		1,
	}
	got := q.ToRoMat()
	if !got.Eq(want) {
		t.Fatalf("ToRoMat is wrong, want: %v, got: %v", want, got)
	}

	dirY := math.Vector{0, 1, 0, 0}
	u = dirY.Unit()
	cosa = math.Cos(angle / 2)
	sina = math.Sin(angle / 2)
	q = math.Quaternion{cosa, math.Vector{sina * u.X, sina * u.Y, sina * u.Z, 0}}
	want = math.Matrix{
		0.5, 0, 0.8660254, 0,
		0, 1, 0, 0,
		-0.8660254, 0, 0.5, 0,
		0, 0, 0, 1,
	}
	got = q.ToRoMat()
	if !got.Eq(want) {
		t.Fatalf("ToRoMat is wrong, want: %v, got: %v", want, got)
	}

	dirZ := math.Vector{0, 0, 1, 0}
	u = dirZ.Unit()
	cosa = math.Cos(angle / 2)
	sina = math.Sin(angle / 2)
	q = math.Quaternion{cosa, math.Vector{sina * u.X, sina * u.Y, sina * u.Z, 0}}
	want = math.Matrix{
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
	dirX := math.Vector{1, 0, 0, 0}
	angle := math.Pi / 3

	u := dirX.Unit()
	cosa := math.Cos(angle / 2)
	sina := math.Sin(angle / 2)
	q := math.NewQuaternion(cosa, sina*u.X, sina*u.Y, sina*u.Z)

	var m math.Matrix
	for i := 0; i < b.N; i++ {
		m = q.ToRoMat()
	}
	_ = m
}
