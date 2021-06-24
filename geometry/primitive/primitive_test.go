// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive_test

import (
	"testing"

	"changkun.de/x/polyred/geometry/primitive"
)

func TestVertex_AABB(t *testing.T) {
	v := primitive.NewRandomVertex()
	pos := v.Pos
	aabb := v.AABB()

	if !aabb.Min.Eq(pos) {
		t.Errorf("vertex aabb min is not euqal to the position")
	}
	if !aabb.Max.Eq(pos) {
		t.Errorf("vertex aabb max is not euqal to the position")
	}
}

func TestTriangle_AABB(t *testing.T) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	aabb1 := primitive.NewTriangle(v1, v2, v3).AABB()
	aabb2 := (&primitive.Triangle{V1: *v1, V2: *v2, V3: *v3}).AABB()
	if !aabb1.Min.Eq(aabb2.Min) {
		t.Errorf("aabb mins are not equal")
	}
	if !aabb1.Max.Eq(aabb2.Max) {
		t.Errorf("aabb mins are not equal")
	}
}

func TestTriangle_FaceNormal(t *testing.T) {
	v1 := primitive.NewRandomVertex()
	v2 := primitive.NewRandomVertex()
	v3 := primitive.NewRandomVertex()

	n1 := primitive.NewTriangle(v1, v2, v3).Normal()
	n2 := (&primitive.Triangle{V1: *v1, V2: *v2, V3: *v3}).Normal()
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
		tri := primitive.Triangle{V1: *v1, V2: *v2, V3: *v3}
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
		tri := primitive.Triangle{V1: *v1, V2: *v2, V3: *v3}
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
