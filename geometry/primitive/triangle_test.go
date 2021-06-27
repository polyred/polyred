// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive_test

import (
	"testing"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/math"
)

func TestTriangle_IsValid(t *testing.T) {

	tri := primitive.NewTriangle(
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
	)

	tri.V1.Pos = math.NewVector(1, 0, 0, 1)
	tri.V2.Pos = math.NewVector(2, 0, 0, 1)
	tri.V3.Pos = math.NewVector(3, 0, 0, 1)

	if tri.IsValid() {
		t.Errorf("invalid triangle returns valid assertion")
	}

	tri.V1.Pos = math.NewVector(1, 0, 0, 1)
	tri.V2.Pos = math.NewVector(2, 0, 0, 1)
	tri.V3.Pos = math.NewVector(0, 1, 0, 1)

	if !tri.IsValid() {
		t.Errorf("valid triangle returns invalid assertion")
	}
}

func TestTriangle_Area(t *testing.T) {
	tri := primitive.NewTriangle(
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
	)
	tri.V1.Pos = math.NewVector(1, 0, 0, 1)
	tri.V2.Pos = math.NewVector(2, 0, 0, 1)
	tri.V3.Pos = math.NewVector(0, 1, 0, 1)

	if tri.Area() != 0.5 {
		t.Error("incorrect triangle area: ", tri.Area())
	}
}

func BenchmarkTriangle_IsValid(b *testing.B) {
	tri := primitive.NewTriangle(
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
	)
	tri.V1.Pos = math.NewVector(1, 0, 0, 1)
	tri.V2.Pos = math.NewVector(2, 0, 0, 1)
	tri.V3.Pos = math.NewVector(0, 1, 0, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tri.IsValid()
	}
}
