// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"strings"
	"testing"

	"poly.red/math"
)

func TestMat_Eq(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		m1 := math.Mat2I[float32]()
		m2 := math.Mat2I[float32]()

		if !m1.Eq(m2) {
			t.Fatalf("unexpected Eq, want true, got false")
		}

		m3 := math.Mat2Zero[float32]()
		if m1.Eq(m3) {
			t.Fatalf("unexpected Eq, want false, got true")
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m1 := math.Mat3I[float32]()
		m2 := math.Mat3I[float32]()

		if !m1.Eq(m2) {
			t.Fatalf("unexpected Eq, want true, got false")
		}

		m3 := math.Mat3Zero[float32]()
		if m1.Eq(m3) {
			t.Fatalf("unexpected Eq, want false, got true")
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m1 := math.Mat4I[float32]()
		m2 := math.Mat4I[float32]()

		if !m1.Eq(m2) {
			t.Fatalf("unexpected Eq, want true, got false")
		}

		m3 := math.Mat4Zero[float32]()
		if m1.Eq(m3) {
			t.Fatalf("unexpected Eq, want false, got true")
		}
	})
}

func TestMat_String(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		m := math.NewMat2[float32](
			1, 2,
			3, 4,
		)

		want := `[
	[1, 2],
	[3, 4],
]`
		t.Log(m)
		if strings.Compare(m.String(), want) != 0 {
			t.Fatalf("string format of Mat2 returns unexpected value, want %v got %v", want, m)
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m := math.NewMat3[float32](
			1, 2, 3,
			4, 5, 6,
			7, 8, 9,
		)

		want := `[
	[1, 2, 3],
	[4, 5, 6],
	[7, 8, 9],
]`
		t.Log(m)
		if strings.Compare(m.String(), want) != 0 {
			t.Fatalf("string format of Mat3 returns unexpected value, want %v got %v", want, m)
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m := math.NewMat4[float32](
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)

		want := `[
	[1, 2, 3, 4],
	[5, 6, 7, 8],
	[9, 10, 11, 12],
	[13, 14, 15, 16],
]`
		t.Log(m)
		if strings.Compare(m.String(), want) != 0 {
			t.Fatalf("string format of Mat4 returns unexpected value, want %v got %v", want, m)
		}
	})
}

func TestMat_Get(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		m := math.NewMat2[float32](
			1, 2,
			3, 4,
		)

		counter := float32(1)
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				if m.Get(i, j) == counter {
					counter++
					continue
				}
				t.Fatalf("unexpected element (%d, %d), got %v, want %v", i, j, m.Get(i, j), counter)
			}
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m := math.NewMat3[float32](
			1, 2, 3,
			4, 5, 6,
			7, 8, 9,
		)

		counter := float32(1)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if m.Get(i, j) == counter {
					counter++
					continue
				}
				t.Fatalf("unexpected element (%d, %d), got %v, want %v", i, j, m.Get(i, j), counter)
			}
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m := math.NewMat4[float32](
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)

		counter := float32(1)
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				if m.Get(i, j) == counter {
					counter++
					continue
				}
				t.Fatalf("unexpected element (%d, %d), got %v, want %v", i, j, m.Get(i, j), counter)
			}
		}
	})

	t.Run("Mat2_Invalid1", func(t *testing.T) {
		m := math.Mat2I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(-1, -1)
	})
	t.Run("Mat2_Invalid2", func(t *testing.T) {
		m := math.Mat2I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(2, 2)
	})

	t.Run("Mat3_Invalid1", func(t *testing.T) {
		m := math.Mat3I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(-1, -1)
	})
	t.Run("Mat3_Invalid2", func(t *testing.T) {
		m := math.Mat3I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(3, 3)
	})

	t.Run("Mat4_Invalid1", func(t *testing.T) {
		m := math.Mat4I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(-1, -1)
	})
	t.Run("Mat4_Invalid2", func(t *testing.T) {
		m := math.Mat4I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Get(4, 4)
	})
}

func TestMat_Set(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		want := math.NewMat2[float32](
			1, 2,
			3, 4,
		)
		got := math.Mat2Zero[float32]()

		counter := float32(1)
		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				got.Set(i, j, counter)
				counter++
			}
		}

		if !want.Eq(got) {
			t.Fatalf("unexpected Set, got %v, want %v", got, want)
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		want := math.NewMat3[float32](
			1, 2, 3,
			4, 5, 6,
			7, 8, 9,
		)
		got := math.Mat3Zero[float32]()

		counter := float32(1)
		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				got.Set(i, j, counter)
				counter++
			}
		}

		if !want.Eq(got) {
			t.Fatalf("unexpected Set, got %v, want %v", got, want)
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		want := math.NewMat4[float32](
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)
		got := math.Mat4Zero[float32]()

		counter := float32(1)
		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				got.Set(i, j, counter)
				counter++
			}
		}

		if !want.Eq(got) {
			t.Fatalf("unexpected Set, got %v, want %v", got, want)
		}
	})

	t.Run("Mat2_Invalid1", func(t *testing.T) {
		m := math.Mat2I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(-1, -1, 1)
	})
	t.Run("Mat2_Invalid2", func(t *testing.T) {
		m := math.Mat2I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(2, 2, 1)
	})

	t.Run("Mat3_Invalid1", func(t *testing.T) {
		m := math.Mat3I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(-1, -1, 1)
	})
	t.Run("Mat3_Invalid2", func(t *testing.T) {
		m := math.Mat3I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(3, 3, 1)
	})

	t.Run("Mat4_Invalid1", func(t *testing.T) {
		m := math.Mat4I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(-1, -1, 1)
	})
	t.Run("Mat4_Invalid2", func(t *testing.T) {
		m := math.Mat4I[float32]()
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("invalid Get does not panic")
			}
		}()
		m.Set(4, 4, 1)
	})
}

func TestMat_Add(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		m1 := math.NewMat2[float32](
			1, 2, 3, 4,
		)
		m2 := math.NewMat2[float32](
			5, 6, 7, 8,
		)
		want := math.NewMat2[float32](
			6, 8, 10, 12,
		)
		if !m1.Add(m2).Eq(want) {
			t.Fatalf("unexpected Mat Add, got %v, want %v", m1.Add(m2), want)
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m1 := math.NewMat3[float32](
			1, 2, 3,
			4, 1, 2,
			3, 4, 1,
		)
		m2 := math.NewMat3[float32](
			5, 6, 7,
			8, 5, 6,
			7, 8, 5,
		)
		want := math.NewMat3[float32](
			6, 8, 10,
			12, 6, 8,
			10, 12, 6,
		)
		if !m1.Add(m2).Eq(want) {
			t.Fatalf("unexpected Mat Add, got %v, want %v", m1.Add(m2), want)
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m1 := math.NewMat4[float32](
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
		)
		m2 := math.NewMat4[float32](
			5, 6, 7, 8,
			5, 6, 7, 8,
			5, 6, 7, 8,
			5, 6, 7, 8,
		)
		want := math.NewMat4[float32](
			6, 8, 10, 12,
			6, 8, 10, 12,
			6, 8, 10, 12,
			6, 8, 10, 12,
		)
		if !m1.Add(m2).Eq(want) {
			t.Fatalf("unexpected Mat Add, got %v, want %v", m1.Add(m2), want)
		}
	})
}

func TestMat_Sub(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		m1 := math.NewMat2[float32](
			1, 2, 3, 4,
		)
		m2 := math.NewMat2[float32](
			5, 6, 7, 8,
		)
		want := math.NewMat2[float32](
			-4, -4, -4, -4,
		)
		if !m1.Sub(m2).Eq(want) {
			t.Fatalf("unexpected Mat Sub, got %v, want %v", m1.Sub(m2), want)
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m1 := math.NewMat3[float32](
			1, 2, 3,
			4, 1, 2,
			3, 4, 1,
		)
		m2 := math.NewMat3[float32](
			5, 6, 7,
			8, 5, 6,
			7, 8, 5,
		)
		want := math.NewMat3[float32](
			-4, -4, -4,
			-4, -4, -4,
			-4, -4, -4,
		)
		if !m1.Sub(m2).Eq(want) {
			t.Fatalf("unexpected Mat Sub, got %v, want %v", m1.Sub(m2), want)
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m1 := math.NewMat4[float32](
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
			1, 2, 3, 4,
		)
		m2 := math.NewMat4[float32](
			5, 6, 7, 8,
			5, 6, 7, 8,
			5, 6, 7, 8,
			5, 6, 7, 8,
		)
		want := math.NewMat4[float32](
			-4, -4, -4, -4,
			-4, -4, -4, -4,
			-4, -4, -4, -4,
			-4, -4, -4, -4,
		)
		if !m1.Sub(m2).Eq(want) {
			t.Fatalf("unexpected Mat Sub, got %v, want %v", m1.Sub(m2), want)
		}
	})
}

func TestMat_MulM(t *testing.T) {
	t.Run("Mat2[float32]", func(t *testing.T) {
		m1 := math.Mat2[float32]{
			1, 2, 3, 4,
		}
		m2 := math.Mat2[float32]{
			16, 15, 14, 13,
		}

		got := m1.MulM(m2)

		want := math.Mat2[float32]{
			44, 41, 104, 97,
		}

		for i := 0; i < 2; i++ {
			for j := 0; j < 2; j++ {
				if got.Get(i, j) == want.Get(i, j) {
					continue
				}
				t.Fatalf("multiply matrices does not working properly, want %v, got %v", want.Get(i, j), got.Get(i, j))
			}
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m1 := math.NewMat3[float32](
			1, 2, 3,
			4, 1, 2,
			3, 4, 1,
		)
		m2 := math.NewMat3[float32](
			16, 15, 14,
			13, 12, 14,
			12, 43, 23,
		)

		got := m1.MulM(m2)

		want := math.Mat3[float32]{
			78, 168, 111, 101, 158, 116, 112, 136, 121,
		}

		for i := 0; i < 3; i++ {
			for j := 0; j < 3; j++ {
				if got.Get(i, j) == want.Get(i, j) {
					continue
				}
				t.Fatalf("multiply matrices does not working properly, want %v, got %v", want.Get(i, j), got.Get(i, j))
			}
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		m1 := math.Mat4[float32]{
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		}
		m2 := math.Mat4[float32]{
			16, 15, 14, 13,
			12, 11, 10, 9,
			8, 7, 6, 5,
			4, 3, 2, 1,
		}

		got := m1.MulM(m2)

		want := math.Mat4[float32]{
			80, 70, 60, 50,
			240, 214, 188, 162,
			400, 358, 316, 274,
			560, 502, 444, 386,
		}

		for i := 0; i < 4; i++ {
			for j := 0; j < 4; j++ {
				if got.Get(i, j) == want.Get(i, j) {
					continue
				}
				t.Fatalf("multiply matrices does not working properly, want %v, got %v", want.Get(i, j), got.Get(i, j))
			}
		}
	})
}

func TestMat_MulV(t *testing.T) {
	t.Run("Mat2", func(t *testing.T) {
		v1 := math.NewVec2[float32](1, 1)
		mat := math.NewMat2[float32](
			1, 2,
			3, 4,
		)
		got := mat.MulV(v1)
		want := math.NewVec2[float32](3, 7)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		v1 := math.NewVec3[float32](1, 1, 1)
		mat := math.NewMat3[float32](
			1, 2, 3,
			4, 5, 6,
			7, 8, 9,
		)
		got := mat.MulV(v1)
		want := math.NewVec3[float32](6, 15, 24)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})

	t.Run("Mat4", func(t *testing.T) {
		v1 := math.NewVec4[float32](1, 1, 1, 1)
		mat := math.NewMat4[float32](
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)
		got := mat.MulV(v1)
		want := math.NewVec4[float32](10, 26, 42, 58)
		if !got.Eq(want) {
			t.Fatalf("unexpected Apply, want %v, got %v", want, got)
		}
	})
}

func TestMat_Det(t *testing.T) {

	t.Run("Mat2", func(t *testing.T) {
		m := math.NewMat2[float32](
			1, 2,
			3, 4,
		)

		want := float32(-2.0)

		if m.Det() != want {
			t.Fatalf("unexpected Det, got %v, want %v", m.Det(), want)
		}

	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m := math.NewMat3[float32](
			1, 2, 3,
			4, 3, 4,
			3, 4, 4,
		)

		want := float32(9.0)

		if m.Det() != want {
			t.Fatalf("unexpected Det, got %v, want %v", m.Det(), want)
		}

	})

	t.Run("Mat4", func(t *testing.T) {
		m := math.Mat4[float32]{
			5, 1, 5, 6,
			8, 2, 2, 3,
			5, 1, 1, 4,
			2, 1, 7, 5,
		}
		want := float32(-44.0)

		if m.Det() != want {
			t.Fatalf("unexpected Det, got %v, want %v", m.Det(), want)
		}
	})
}

func TestMat_T(t *testing.T) {

	t.Run("Mat2", func(t *testing.T) {
		m := math.NewMat2[float32](
			1, 2,
			3, 4,
		)

		want := math.NewMat2[float32](
			1, 3,
			2, 4,
		)

		if !m.T().Eq(want) {
			t.Fatalf("unexpected T, got %v, want %v", m.T(), want)
		}

	})

	t.Run("Mat3[float32]", func(t *testing.T) {
		m := math.NewMat3[float32](
			1, 2, 3,
			4, 3, 4,
			3, 4, 4,
		)

		want := math.NewMat3[float32](
			1, 4, 3,
			2, 3, 4,
			3, 4, 4,
		)

		if !m.T().Eq(want) {
			t.Fatalf("unexpected T, got %v, want %v", m.T(), want)
		}

	})

	t.Run("Mat4", func(t *testing.T) {
		m := math.Mat4[float32]{
			5, 1, 5, 6,
			8, 2, 2, 3,
			5, 1, 1, 4,
			2, 1, 7, 5,
		}
		want := math.Mat4[float32]{
			5, 8, 5, 2,
			1, 2, 1, 1,
			5, 2, 1, 7,
			6, 3, 4, 5,
		}

		if !m.T().Eq(want) {
			t.Fatalf("unexpected T, got %v, want %v", m.T(), want)
		}
	})
}

func TestMat_Inv(t *testing.T) {
	t.Run("Mat4", func(t *testing.T) {
		m := math.Mat4[float32]{
			5, 1, 5, 6,
			8, 71, 2, 47,
			5, 1, 582, 4,
			2, 1, 7, 25,
		}
		m = m.Inv()

		want := math.Mat4[float32]{
			1003995.0 / 4463716, -10967.0 / 4463716, -5949.0 / 4463716, -219389.0 / 4463716,
			-62879.0 / 4463716, 65251.0 / 4463716, 1613.0 / 4463716, -107839.0 / 4463716,
			-3999.0 / 2231858, -3.0 / 2231858, 3865.0 / 2231858, 347.0 / 2231858,
			-75565.0 / 4463716, -1731.0 / 4463716, -1753.0 / 4463716, 200219.0 / 4463716,
		}

		if m.Eq(want) {
			return
		}

		t.Fatalf("unexpected Inv, got %+v, want %+v", m, want)
	})

	t.Run("Mat4_Invalid", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("zero matrix inverse should panic")
			}
		}()
		m := math.Mat4Zero[float32]()
		m = m.Inv()
	})
}
