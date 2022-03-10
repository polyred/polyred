// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"image/color"
	"testing"

	"poly.red/math"
)

var (
	global  float32
	globalV math.Vec4[float32]
	globalC color.RGBA
)

func TestLerp(t *testing.T) {
	t.Run("float64", func(t *testing.T) {
		if v := math.Lerp[float32](0, 1, 0.5); v != 0.5 {
			t.Fatalf("Lerp want %v, got %v", 0.5, v)
		}
	})
	t.Run("int", func(t *testing.T) {
		if v := math.LerpInt(0, 2, 1); v != 2 {
			t.Fatalf("Lerp want %v, got %v", 2, v)
		}
	})
}

func BenchmarkLerp(b *testing.B) {
	t := float32(0.5)
	var r float32
	for i := 0; i < b.N; i++ {
		r = math.Lerp(0, 1, t)
	}
	global = r
}

func TestLerpVec(t *testing.T) {
	t.Run("Vec4", func(t *testing.T) {
		tt := float32(0.5)
		v1 := math.Vec4[float32]{0, 0, 0, 1}
		v2 := math.Vec4[float32]{1, 1, 1, 1}
		want := math.Vec4[float32]{0.5, 0.5, 0.5, 1}
		if vv := math.LerpVec4(v1, v2, tt); !vv.Eq(want) {
			t.Fatalf("LerpVec4 want %v, got %v", want, vv)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		tt := float32(0.5)
		v1 := math.Vec3[float32]{0, 0, 0}
		v2 := math.Vec3[float32]{1, 1, 1}
		want := math.Vec3[float32]{0.5, 0.5, 0.5}
		if vv := math.LerpVec3(v1, v2, tt); !vv.Eq(want) {
			t.Fatalf("LerpVec3 want %v, got %v", want, vv)
		}
	})
	t.Run("Vec2", func(t *testing.T) {
		tt := float32(0.5)
		v1 := math.Vec2[float32]{0, 0}
		v2 := math.Vec2[float32]{1, 1}
		want := math.Vec2[float32]{0.5, 0.5}
		if vv := math.LerpVec2(v1, v2, tt); !vv.Eq(want) {
			t.Fatalf("LerpVec2 want %v, got %v", want, vv)
		}
	})
}

func BenchmarkLerpV(b *testing.B) {
	t := float32(0.5)
	v1 := math.Vec4[float32]{0, 0, 0, 1}
	v2 := math.Vec4[float32]{1, 1, 1, 1}
	var r math.Vec4[float32]
	for i := 0; i < b.N; i++ {
		r = math.LerpVec4(v1, v2, t)
	}
	globalV = r
}

func TestLerpC(t *testing.T) {
	tt := float32(0.5)
	v1 := color.RGBA{0, 0, 0, 255}
	v2 := color.RGBA{255, 255, 255, 255}
	want := color.RGBA{127, 127, 127, 255}
	if vv := math.LerpC(v1, v2, tt); vv.R != want.R || vv.G != want.G || vv.B != want.B || vv.A != want.A {
		t.Fatalf("LerpC want %v, got %v", want, vv)
	}
}

func BenchmarkLerpC(b *testing.B) {
	t := float32(0.5)
	v1 := color.RGBA{0, 0, 0, 255}
	v2 := color.RGBA{255, 255, 255, 255}
	var r color.RGBA
	for i := 0; i < b.N; i++ {
		r = math.LerpC(v1, v2, t)
	}
	globalC = r
}

func TestBarycoord(t *testing.T) {
	p := math.Vec2[float32]{5., 5.}
	v1 := math.NewVec2[float32](0, 0)
	v2 := math.NewVec2[float32](20, 0)
	v3 := math.NewVec2[float32](0, 20)

	barycoords := math.Barycoord(p, v1, v2, v3)

	if barycoords[0] != 0.5 || barycoords[1] != 0.25 || barycoords[2] != 0.25 {
		t.Fatalf("barycentric coordinates does not match: %v", barycoords)
	}
}

func BenchmarkBarycoord(b *testing.B) {
	v1 := math.NewVec2[float32](0, 0)
	v2 := math.NewVec2[float32](20, 0)
	v3 := math.NewVec2[float32](0, 20)
	b.Run("inside", func(b *testing.B) {
		p := math.Vec2[float32]{10, 10}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = math.Barycoord(p, v1, v2, v3)
		}
		b.StopTimer()
	})
	b.Run("outside", func(b *testing.B) {
		p := math.Vec2[float32]{20, 20}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = math.Barycoord(p, v1, v2, v3)
		}
		b.StopTimer()
	})
}

func TestIsInsideTriangle(t *testing.T) {
	v1 := math.NewVec2[float32](0, 0)
	v2 := math.NewVec2[float32](20, 0)
	v3 := math.NewVec2[float32](0, 20)
	t.Run("inside", func(t *testing.T) {
		p := math.Vec2[float32]{10, 10}
		if !math.IsInsideTriangle(p, v1, v2, v3) {
			t.Fatalf("unexpected inside triangle test, want true, got false")
		}
	})
	t.Run("outside", func(t *testing.T) {
		p := math.Vec2[float32]{20, 20}
		if math.IsInsideTriangle(p, v1, v2, v3) {
			t.Fatalf("unexpected inside triangle test, want false, got true")
		}

		p = math.Vec2[float32]{-20, -20}
		if math.IsInsideTriangle(p, v1, v2, v3) {
			t.Fatalf("unexpected inside triangle test, want false, got true")
		}
	})
}

func BenchmarkIsInsideTriangle(b *testing.B) {
	v1 := math.NewVec2[float32](0, 0)
	v2 := math.NewVec2[float32](20, 0)
	v3 := math.NewVec2[float32](0, 20)
	b.Run("inside", func(b *testing.B) {
		p := math.Vec2[float32]{10, 10}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = math.IsInsideTriangle(p, v1, v2, v3)
		}
		b.StopTimer()
	})
	b.Run("outside", func(b *testing.B) {
		p := math.Vec2[float32]{20, 20}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = math.IsInsideTriangle(p, v1, v2, v3)
		}
		b.StopTimer()
	})
}
