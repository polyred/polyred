// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive_test

import (
	"testing"

	"poly.red/geometry/primitive"
	"poly.red/math"
)

func TestTriangle_AABB(t *testing.T) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	aabb1 := primitive.NewTriangle(v1, v2, v3).AABB()
	aabb2 := (&primitive.Triangle{V1: v1, V2: v2, V3: v3}).AABB()
	if !aabb1.Min.Eq(aabb2.Min) {
		t.Errorf("aabb mins are not equal")
	}
	if !aabb1.Max.Eq(aabb2.Max) {
		t.Errorf("aabb mins are not equal")
	}
}

func TestTriangle_IsValid(t *testing.T) {

	tri := primitive.NewTriangle(
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
	)

	tri.V1.Pos = math.NewVec4[float32](1, 0, 0, 1)
	tri.V2.Pos = math.NewVec4[float32](2, 0, 0, 1)
	tri.V3.Pos = math.NewVec4[float32](3, 0, 0, 1)

	if tri.IsValid() {
		t.Errorf("invalid triangle returns valid assertion")
	}

	tri.V1.Pos = math.NewVec4[float32](1, 0, 0, 1)
	tri.V2.Pos = math.NewVec4[float32](2, 0, 0, 1)
	tri.V3.Pos = math.NewVec4[float32](0, 1, 0, 1)

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
	tri.V1.Pos = math.NewVec4[float32](1, 0, 0, 1)
	tri.V2.Pos = math.NewVec4[float32](2, 0, 0, 1)
	tri.V3.Pos = math.NewVec4[float32](0, 1, 0, 1)

	if tri.Area() != 0.5 {
		t.Error("incorrect triangle area: ", tri.Area())
	}
}

func TestTriangle_FaceNormal(t *testing.T) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	n1 := primitive.NewTriangle(v1, v2, v3).Normal()
	n2 := (&primitive.Triangle{V1: v1, V2: v2, V3: v3}).Normal()
	if !n1.Eq(n2) {
		t.Errorf("face normals are not equal")
	}
	if !n1.Eq(n2) {
		t.Errorf("face normals are not equal")
	}
}

func BenchmarkTriangle_AABB(b *testing.B) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	b.Run("plain", func(b *testing.B) {
		tri := primitive.Triangle{V1: v1, V2: v2, V3: v3}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tri.AABB()
		}
	})

	b.Run("alloc", func(b *testing.B) {
		tri := primitive.NewTriangle(v1, v2, v3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tri.AABB()
		}
	})
}

func BenchmarkTriangle_FaceNormal(b *testing.B) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	b.Run("plain", func(b *testing.B) {
		tri := primitive.Triangle{V1: v1, V2: v2, V3: v3}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tri.Normal()
		}
	})

	b.Run("alloc", func(b *testing.B) {
		tri := primitive.NewTriangle(v1, v2, v3)
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tri.Normal()
		}
	})
}

func BenchmarkTriangle_IsValid(b *testing.B) {
	tri := primitive.NewTriangle(
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
		primitive.NewRandomVertex(),
	)
	tri.V1.Pos = math.NewVec4[float32](1, 0, 0, 1)
	tri.V2.Pos = math.NewVec4[float32](2, 0, 0, 1)
	tri.V3.Pos = math.NewVec4[float32](0, 1, 0, 1)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tri.IsValid()
	}
}
