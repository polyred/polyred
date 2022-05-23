// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"strings"
	"testing"

	"poly.red/math"
)

func TestVec_New(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		v2 := math.NewVec2[float32](2, 2)
		v3 := math.NewVec2[float32](1, 1)
		if v1.Eq(v2) {
			t.Fatalf("unexpected comparison, got true, want false")
		}
		if !v1.Eq(v3) {
			t.Fatalf("unexpected comparison, got false, want true")
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 2)
		v2 := math.NewVec3[float32](2, 2, 2)
		v3 := math.NewVec3[float32](1, 1, 2)
		if v1.Eq(v2) {
			t.Fatalf("unexpected comparison, got true, want false")
		}
		if !v1.Eq(v3) {
			t.Fatalf("unexpected comparison, got false, want true")
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 4, 2)
		v2 := math.NewVec4[float32](2, 2, 4, 2)
		v3 := math.NewVec4[float32](1, 1, 4, 2)
		if v1.Eq(v2) {
			t.Fatalf("unexpected comparison, got true, want false")
		}
		if !v1.Eq(v3) {
			t.Fatalf("unexpected comparison, got false, want true")
		}
	})
}

func TestVec_Rand(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewRandVec2[float32]()
		v2 := math.NewRandVec2[float32]()
		if v1.Eq(v2) {
			t.Fatalf("unexpected different random vectors, got %v, %v", v1, v2)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewRandVec3[float32]()
		v2 := math.NewRandVec3[float32]()
		if v1.Eq(v2) {
			t.Fatalf("unexpected different random vectors, got %v, %v", v1, v2)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewRandVec4[float32]()
		v2 := math.NewRandVec4[float32]()
		if v1.Eq(v2) {
			t.Fatalf("unexpected different random vectors, got %v, %v", v1, v2)
		}
	})
}

func TestVec_String(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v := math.NewVec2[float32](1, 2)
		want := "<1, 2>"
		t.Log(v)
		if strings.Compare(v.String(), want) != 0 {
			t.Fatalf("unexpected String, got %v, want %v", v.String(), want)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v := math.NewVec3[float32](1, 2, 3)
		want := "<1, 2, 3>"
		t.Log(v)
		if strings.Compare(v.String(), want) != 0 {
			t.Fatalf("unexpected String, got %v, want %v", v.String(), want)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v := math.NewVec4[float32](1, 2, 3, 4)
		want := "<1, 2, 3, 4>"
		t.Log(v)
		if strings.Compare(v.String(), want) != 0 {
			t.Fatalf("unexpected String, got %v, want %v", v.String(), want)
		}
	})
}

func TestVec_Add(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		v2 := math.NewVec2[float32](2, 2)
		got := v1.Add(v2)
		want := math.NewVec2[float32](3, 3)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 3)
		v2 := math.NewVec3[float32](2, 2, 3)
		got := v1.Add(v2)
		want := math.NewVec3[float32](3, 3, 6)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 2, 5)
		v2 := math.NewVec4[float32](2, 2, 2, 5)
		got := v1.Add(v2)
		want := math.NewVec4[float32](3, 3, 4, 10)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
}

func TestVec_Sub(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		v2 := math.NewVec2[float32](2, 2)
		got := v1.Sub(v2)
		want := math.NewVec2[float32](-1, -1)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 3)
		v2 := math.NewVec3[float32](2, 2, 3)
		got := v1.Sub(v2)
		want := math.NewVec3[float32](-1, -1, 0)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 2, 5)
		v2 := math.NewVec4[float32](2, 2, 2, 5)
		got := v1.Sub(v2)
		want := math.NewVec4[float32](-1, -1, 0, 0)
		if !want.Eq(got) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
}

func TestVec_Dot(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		v2 := math.NewVec2[float32](2, 2)
		got := v1.Dot(v2)
		want := float32(4.0)
		if !math.ApproxEq(want, got, math.Epsilon) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 3)
		v2 := math.NewVec3[float32](2, 2, 3)
		got := v1.Dot(v2)
		want := float32(13.0)
		if !math.ApproxEq(want, got, math.Epsilon) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 2, 5)
		v2 := math.NewVec4[float32](2, 2, 2, 5)
		got := v1.Dot(v2)
		want := float32(33.0)
		if !math.ApproxEq(want, got, math.Epsilon) {
			t.Fatalf("unexpected Add, got %v, want %v", got, want)
		}
	})
}

func TestVec_IsZero(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		if v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want false, got true")
		}
		v1 = math.NewVec2[float32](0, 0)
		if !v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want true, got false")
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 3)
		if v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want false, got true")
		}
		v1 = math.NewVec3[float32](0, 0, 0)
		if !v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want true, got false")
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 2, 5)
		if v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want false, got true")
		}
		v1 = math.NewVec4[float32](0, 0, 0, 0)
		if !v1.IsZero() {
			t.Fatalf("unexpected IsZero assertion, want true, got false")
		}
	})
}

func TestVec_Scale(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		got := v1.Scale(2, 2)
		want := math.NewVec2[float32](2, 2)
		if !got.Eq(want) {
			t.Fatalf("unexpected scale, want %v, got %v", want, got)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		got := v1.Scale(2, 2, 2)
		want := math.NewVec3[float32](2, 2, 2)
		if !got.Eq(want) {
			t.Fatalf("unexpected scale, want %v, got %v", want, got)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 1)
		got := v1.Scale(2, 2, 2, 2)
		want := math.NewVec4[float32](2, 2, 2, 2)
		if !got.Eq(want) {
			t.Fatalf("unexpected scale, want %v, got %v", want, got)
		}
	})
}

func TestVec_Translate(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		got := v1.Translate(2, 2)
		want := math.NewVec2[float32](3, 3)
		if !got.Eq(want) {
			t.Fatalf("unexpected translate, want %v, got %v", want, got)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		got := v1.Translate(2, 2, 2)
		want := math.NewVec3[float32](3, 3, 3)
		if !got.Eq(want) {
			t.Fatalf("unexpected translate, want %v, got %v", want, got)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 1)
		got := v1.Translate(2, 2, 2)
		want := math.NewVec4[float32](3, 3, 3, 1)
		if !got.Eq(want) {
			t.Fatalf("unexpected translate, want %v, got %v", want, got)
		}

		v2 := math.NewVec4[float32](1, 1, 1, 0)
		got = v2.Translate(2, 2, 2)
		want = math.NewVec4[float32](1, 1, 1, 0)
		if !got.Eq(want) {
			t.Fatalf("unexpected translate, want %v, got %v", want, got)
		}
	})
}

func TestVec_Len(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		got := v1.Len()
		want := math.Sqrt[float32](2)
		if !math.ApproxEq(got, want, math.Epsilon) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		got := v1.Len()
		want := math.Sqrt[float32](3)
		if !math.ApproxEq(got, want, math.Epsilon) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})
	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 0)
		got := v1.Len()
		want := math.Sqrt[float32](3.0)
		if !math.ApproxEq(got, want, math.Epsilon) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}

		v1 = math.NewVec4[float32](1, 1, 1, 1)
		got = v1.Len()
		want = 2
		if !math.ApproxEq(got, want, math.Epsilon) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})
}

func TestVec_Unit(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		got := v1.Unit()
		want := math.NewVec2(1/math.Sqrt[float32](2), 1/math.Sqrt[float32](2))
		if !got.Eq(want) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		got := v1.Unit()
		want := math.NewVec3(1/math.Sqrt[float32](3), 1/math.Sqrt[float32](3), 1/math.Sqrt[float32](3))
		if !got.Eq(want) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})

	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 0)
		got := v1.Unit()
		want := math.NewVec4(1/math.Sqrt[float32](3), 1/math.Sqrt[float32](3), 1/math.Sqrt[float32](3), 0)
		if !got.Eq(want) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}

		v1 = math.NewVec4[float32](1, 1, 1, 1)
		got = v1.Unit()
		want = math.NewVec4[float32](0.5, 0.5, 0.5, 0.5)
		if !got.Eq(want) {
			t.Fatalf("unexpected Len, want %v, got %v", want, got)
		}
	})
}

func TestVec_Apply(t *testing.T) {
	t.Run("Vec2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		mat := math.NewMat2[float32](
			1, 2,
			3, 4,
		)
		got := v1.Apply(mat)
		want := math.NewVec2[float32](3, 7)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})

	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		mat := math.NewMat3[float32](
			1, 2, 3,
			4, 5, 6,
			7, 8, 9,
		)
		got := v1.Apply(mat)
		want := math.NewVec3[float32](6, 15, 24)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})

	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 1)
		mat := math.NewMat4[float32](
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)
		got := v1.Apply(mat)
		want := math.NewVec4[float32](10, 26, 42, 58)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})
}

func TestVec_Cross(t *testing.T) {
	t.Run("Vec3", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 0, 0)
		v2 := math.NewVec3[float32](0, 1, 0)
		want := math.NewVec3[float32](0, 0, 1)
		got := v1.Cross(v2)
		if !got.Eq(want) {
			t.Fatalf("unexpected Cross, want %v, got %v", want, got)
		}
	})

	t.Run("Vec4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 0, 0, 0)
		v2 := math.NewVec4[float32](0, 1, 0, 0)
		want := math.NewVec4[float32](0, 0, 1, 0)
		got := v1.Cross(v2)
		if !got.Eq(want) {
			t.Fatalf("unexpected Cross, want %v, got %v", want, got)
		}
	})
}

func TestVec_Convert(t *testing.T) {

	t.Run("Vec3ToVec4", func(t *testing.T) {
		v1 := math.NewRandVec3[float32]()
		got := v1.ToVec4(1)
		want := math.NewVec4(v1.X, v1.Y, v1.Z, 1)
		if !got.Eq(want) {
			t.Fatalf("unexpected Vec3ToVec4, got %v, want %v", got, want)
		}
	})
	t.Run("Vec4ToVec3", func(t *testing.T) {
		v1 := math.NewRandVec4[float32]()
		got := v1.ToVec3()
		want := math.NewVec3(v1.X, v1.Y, v1.Z)
		if !got.Eq(want) {
			t.Fatalf("unexpected Vec4ToVec3, got %v, want %v", got, want)
		}
	})
	t.Run("Vec4ToVec2", func(t *testing.T) {
		v1 := math.NewRandVec4[float32]()
		got := v1.ToVec2()
		want := math.NewVec2(v1.X, v1.Y)
		if !got.Eq(want) {
			t.Fatalf("unexpected Vec4ToVec3, got %v, want %v", got, want)
		}
	})

	t.Run("Vec4_Pos", func(t *testing.T) {
		v1 := math.NewRandVec3[float32]()
		got := v1.ToVec4(2).Pos()
		want := v1.Scale(0.5, 0.5, 0.5).ToVec4(1)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}

		v2 := math.NewVec4[float32](1, 1, 1, 0)
		got = v2.Pos()
		want = math.NewVec4[float32](1, 1, 1, 1)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}

		v2 = math.NewVec4[float32](1, 1, 1, 1)
		got = v2.Pos()
		want = math.NewVec4[float32](1, 1, 1, 1)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}
	})

	t.Run("Vec4_Vec", func(t *testing.T) {
		v1 := math.NewRandVec3[float32]()
		got := v1.ToVec4(2).Vec()
		want := v1.Scale(0.5, 0.5, 0.5).ToVec4(0)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}

		v2 := math.NewVec4[float32](1, 1, 1, 0)
		got = v2.Vec()
		want = math.NewVec4[float32](1, 1, 1, 0)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}

		v2 = math.NewVec4[float32](1, 1, 1, 1)
		got = v2.Vec()
		want = math.NewVec4[float32](1, 1, 1, 0)
		if !got.Eq(want) {
			t.Fatalf("unexpected Pos, got %v, want %v", got, want)
		}
	})
}

func FuzzVec_Add(f *testing.F) {
	f.Add(float32(1), float32(1), float32(2), float32(3))
	f.Fuzz(func(t *testing.T, a1, a2, a3, a4 float32) {
		v1 := math.NewVec2(a1, a2)
		v2 := math.NewVec2(a3, a4)
		want := math.NewVec2(a1+a3, a2+a4)
		if !v1.Add(v2).Eq(want) {
			t.Fatalf("unexpected add results: %v+%v=%v, want: %v", v1, v2, v1.Add(v2), want)
		}
	})
}
