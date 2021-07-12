// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"testing"

	"changkun.de/x/polyred/math"
)

func TestClamp(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		got := math.Clamp(128, 0, 255)
		if got != 128 {
			t.Fatalf("unexpected clamp, got %v, want 128", got)
		}

		got = math.Clamp(-1, 0, 255)
		if got != 0 {
			t.Fatalf("unexpected clamp, got %v, want 0", got)
		}

		got = math.Clamp(256, 0, 255)
		if got != 255 {
			t.Fatalf("unexpected clamp, got %v, want 255", got)
		}
	})

	t.Run("int", func(t *testing.T) {
		got := math.ClampInt(128, 0, 255)
		if got != 128 {
			t.Fatalf("unexpected clamp, got %v, want 128", got)
		}

		got = math.ClampInt(-1, 0, 255)
		if got != 0 {
			t.Fatalf("unexpected clamp, got %v, want 0", got)
		}

		got = math.ClampInt(256, 0, 255)
		if got != 255 {
			t.Fatalf("unexpected clamp, got %v, want 255", got)
		}
	})

	t.Run("Vec4", func(t *testing.T) {
		v := math.Vec4{128, 128, 128, 128}
		want := math.Vec4{128, 128, 128, 128}
		got := math.ClampVec4(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec4{-1, -1, -1, -1}
		want = math.Vec4{0, 0, 0, 0}
		got = math.ClampVec4(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec4{256, 266, 265, 256}
		want = math.Vec4{255, 255, 255, 255}
		got = math.ClampVec4(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want 2%v55", got, want)
		}
	})

	t.Run("Vec3", func(t *testing.T) {
		v := math.Vec3{128, 128, 128}
		want := math.Vec3{128, 128, 128}
		got := math.ClampVec3(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec3{-1, -1, -1}
		want = math.Vec3{0, 0, 0}
		got = math.ClampVec3(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec3{256, 266, 265}
		want = math.Vec3{255, 255, 255}
		got = math.ClampVec3(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want 2%v55", got, want)
		}
	})

	t.Run("Vec2", func(t *testing.T) {
		v := math.Vec2{128, 128}
		want := math.Vec2{128, 128}
		got := math.ClampVec2(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec2{-1, -1}
		want = math.Vec2{0, 0}
		got = math.ClampVec2(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want %v", got, want)
		}

		v = math.Vec2{256, 256}
		want = math.Vec2{255, 255}
		got = math.ClampVec2(v, 0, 255)
		if got != want {
			t.Fatalf("unexpected clamp, got %v, want 2%v55", got, want)
		}
	})
}

func BenchmarkClamp(b *testing.B) {
	v := 128.0

	var bb float64
	for i := 0; i < b.N; i++ {
		bb = math.Clamp(v, 0, 255)
	}
	_ = bb
}

func BenchmarkClampInt(b *testing.B) {
	v := 128

	var bb int
	for i := 0; i < b.N; i++ {
		bb = math.ClampInt(v, 0, 255)
	}
	_ = bb
}

func BenchmarkClampVec(b *testing.B) {
	b.Run("Vec4", func(b *testing.B) {
		v := math.Vec4{128, 128, 128, 255}

		var bb math.Vec4
		for i := 0; i < b.N; i++ {
			bb = math.ClampVec4(v, 0, 255)
		}
		_ = bb
	})

	b.Run("Vec3", func(b *testing.B) {
		v := math.Vec3{128, 128, 128}

		var bb math.Vec3
		for i := 0; i < b.N; i++ {
			bb = math.ClampVec3(v, 0, 255)
		}
		_ = bb
	})

	b.Run("Vec2", func(b *testing.B) {
		v := math.Vec2{128, 128}

		var bb math.Vec2
		for i := 0; i < b.N; i++ {
			bb = math.ClampVec2(v, 0, 255)
		}
		_ = bb
	})
}
