// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"strings"
	"testing"

	"poly.red/math"
)

func TestMat(t *testing.T) {
	t.Run("MatN", func(t *testing.T) {
		m := math.NewMat[float32](
			5, 5,
			1, 2, 3, 4, 5,
			6, 7, 8, 9, 10,
			11, 12, 13, 14, 15,
			16, 17, 18, 19, 20,
			21, 22, 23, 24, 25,
		)

		want := `[
	[1, 2, 3, 4, 5],
	[6, 7, 8, 9, 10],
	[11, 12, 13, 14, 15],
	[16, 17, 18, 19, 20],
	[21, 22, 23, 24, 25],
]`
		t.Log(m)
		if strings.Compare(m.String(), want) != 0 {
			t.Fatalf("string format of Mat returns unexpected value, want %v got %v", want, m)
		}
	})
	t.Run("Mat_Get", func(t *testing.T) {
		m := math.NewMat[float32](
			2, 3,
			1, 2, 3,
			4, 5, 6,
		)

		assert(t, m.Get(0, 0), 1)
		assert(t, m.Get(0, 1), 2)
		assert(t, m.Get(0, 2), 3)
		assert(t, m.Get(1, 0), 4)
		assert(t, m.Get(1, 1), 5)
		assert(t, m.Get(1, 2), 6)

		t.Log(m)
	})

	t.Run("Mat_Mul", func(t *testing.T) {
		m1 := math.NewMat[float32](
			4, 4,
			1, 2, 3, 4,
			5, 6, 7, 8,
			9, 10, 11, 12,
			13, 14, 15, 16,
		)
		m2 := math.NewMat[float32](
			4, 4,
			16, 15, 14, 13,
			12, 11, 10, 9,
			8, 7, 6, 5,
			4, 3, 2, 1,
		)

		got := m1.Mul(m2)
		want := math.NewMat[float32](
			4, 4,
			80, 70, 60, 50,
			240, 214, 188, 162,
			400, 358, 316, 274,
			560, 502, 444, 386,
		)

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

func assert[T comparable](t *testing.T, want T, got T) {
	if want == got {
		return
	}
	t.Fatalf("want %v, got %v", want, got)
}

var mm math.Mat[float32]

func BenchmarkMat_Mul(b *testing.B) {
	m1 := math.NewRandMat[float32](100, 100)
	m2 := math.NewRandMat[float32](100, 100)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		mm = m1.Mul(m2)
	}
}
