// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive_test

import (
	"testing"

	"poly.red/geometry/primitive"
	"poly.red/math"
)

func TestNewAABB(t *testing.T) {
	v1 := math.NewVec3[float32](1, 0, 0)
	v2 := math.NewVec3[float32](0, 1, 0)
	v3 := math.NewVec3[float32](0, 0, 1)

	aabb := primitive.NewAABB(v1, v2, v3)

	if !aabb.Min.Eq(math.NewVec3[float32](0, 0, 0)) {
		t.Fatal("not equal")
	}
	if !aabb.Max.Eq(math.NewVec3[float32](1, 1, 1)) {
		t.Fatal("not equal")
	}

}

func TestAABB_Intersect(t *testing.T) {

	v1 := math.NewVec3[float32](1, 0, 0)
	v2 := math.NewVec3[float32](0, 1, 0)
	v3 := math.NewVec3[float32](0, 0, 1)

	aabb1 := primitive.NewAABB(v1, v2, v3)

	v4 := math.NewVec3[float32](-1, -0.5, -0.5)
	v5 := math.NewVec3[float32](-0.5, -1, -0.5)
	v6 := math.NewVec3[float32](-0.5, -0.5, -1)

	aabb2 := primitive.NewAABB(v4, v5, v6)

	if aabb1.Intersect(aabb2) {
		t.Fatalf("intersect")
	}
	v7 := math.NewVec3[float32](0.5, 0, 0)
	v8 := math.NewVec3[float32](0, 0.5, 0)
	v9 := math.NewVec3[float32](0, 0, 0.5)

	aabb3 := primitive.NewAABB(v7, v8, v9)

	if !aabb1.Intersect(aabb3) {
		t.Fatalf("not intersect")
	}

	v10 := math.NewVec3[float32](-1, 0, 0)
	v11 := math.NewVec3[float32](0, -1, 0)
	v12 := math.NewVec3[float32](0, 0, -1)

	aabb4 := primitive.NewAABB(v10, v11, v12)

	if !aabb1.Intersect(aabb4) {
		t.Fatalf("not intersect")
	}
}

func TestAABB_Add(t *testing.T) {

	v1 := math.NewVec3[float32](1, 0, 0)
	v2 := math.NewVec3[float32](0, 1, 0)
	v3 := math.NewVec3[float32](0, 0, 1)

	aabb := primitive.NewAABB(v1, v2, v3)

	v4 := math.NewVec3[float32](-1, -0.5, -0.5)
	v5 := math.NewVec3[float32](-0.5, -1, -0.5)
	v6 := math.NewVec3[float32](-0.5, -0.5, -1)

	aabb.Add(primitive.NewAABB(v4, v5, v6))
	want := primitive.NewAABB(v1, v2, v3, v4, v5, v6)
	if !aabb.Eq(want) {
		t.Fatalf("AABB add does not work")
	}
}

func BenchmarkVertexAABB(b *testing.B) {
	v := primitive.NewRandomVertex()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v.AABB()
	}
}

func TestAABB_Contains(t *testing.T) {
	v1 := math.NewVec3[float32](1, 0, 0)
	v2 := math.NewVec3[float32](0, 1, 0)
	v3 := math.NewVec3[float32](0, 0, 1)
	aabb := primitive.NewAABB(v1, v2, v3)

	if !aabb.Contains(v1, v2, v3) {
		t.Fatalf("AABB should contains their boundary points, but actually not.")
	}

	if aabb.Contains(math.NewVec3[float32](-1, -1, -1)) {
		t.Fatalf("AABB should not contain outside points, but actually contained.")
	}
}
