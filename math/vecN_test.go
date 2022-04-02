// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"strings"
	"testing"

	"poly.red/math"
)

func TestVec(t *testing.T) {
	t.Run("NewVec", func(t *testing.T) {
		v1 := math.NewVec[float32](1, 1, 4, 2, 2)
		v2 := math.NewVec[float32](2, 2, 4, 2, 2)
		v3 := math.NewVec[float32](1, 1, 4, 2, 2)
		if v1.Eq(v2) {
			t.Fatalf("unexpected comparison, got true, want false")
		}
		if !v1.Eq(v3) {
			t.Fatalf("unexpected comparison, got false, want true")
		}
	})

	t.Run("Vec_String", func(t *testing.T) {
		v := math.NewVec[float32](1, 2, 3, 4, 5)
		want := "<1, 2, 3, 4, 5>"
		t.Log(v)
		if strings.Compare(v.String(), want) != 0 {
			t.Fatalf("unexpected String, got %v, want %v", v.String(), want)
		}
	})
}
