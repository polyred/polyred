// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd_test

import (
	"testing"

	"changkun.de/x/ddd"
)

func TestVectorAddition(t *testing.T) {
	v1 := ddd.Vector{1, 2, 3, 4}
	v2 := ddd.Vector{4, 5, 6, 7}

	v1 = v1.Add(v2)
	want := ddd.Vector{5, 7, 9, 11}

	if !equal(want, v1) {
		t.Fatalf("vector addtion is not working, want %v, got %v", want, v1)
	}
}

func equal(v1, v2 ddd.Vector) bool {
	if v1.X == v2.X && v1.Y == v2.Y && v1.Z == v2.Z && v1.W == v2.W {
		return true
	}
	return false
}
