// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd_test

import (
	"math"
	"testing"

	"github.com/changkun/ddd"
)

func TestNewMatrix(t *testing.T) {
	m := ddd.IdentityMatrix

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if i == j && m.Get(i, j) == 1 {
				continue
			}
			if i != j && m.Get(i, j) == 0 {
				continue
			}
			t.Fatalf("new matrix is not an intentity matrix")
		}
	}
}

func TestSetMatrix(t *testing.T) {
	m := ddd.Matrix{
		1, 1, 1, 1,
		1, 1, 1, 1,
		1, 1, 1, 1,
		1, 1, 1, 1,
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if m.Get(i, j) == 1 {
				continue
			}
			t.Fatalf("set matrix does not working properly, want 1, got %v", m.Get(i, j))
		}
	}
}

func TestMultiplyMatrices(t *testing.T) {
	m1 := ddd.Matrix{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	m2 := ddd.Matrix{
		16, 15, 14, 13,
		12, 11, 10, 9,
		8, 7, 6, 5,
		4, 3, 2, 1,
	}

	got := m1.Mul(m2)

	want := ddd.Matrix{
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
}

func TestInverseMatrix(t *testing.T) {
	m1 := ddd.Matrix{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m1 = m1.Inverse()

	want := ddd.Matrix{
		1003995.0 / 4463716, -10967.0 / 4463716, -5949.0 / 4463716, -219389.0 / 4463716,
		-62879.0 / 4463716, 65251.0 / 4463716, 1613.0 / 4463716, -107839.0 / 4463716,
		-3999.0 / 2231858, -3.0 / 2231858, 3865.0 / 2231858, 347.0 / 2231858,
		-75565.0 / 4463716, -1731.0 / 4463716, -1753.0 / 4463716, 200219.0 / 4463716,
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if math.Abs(m1.Get(i, j)-want.Get(i, j)) < 1e-5 {
				continue
			}
			t.Fatalf("inverse matrices does not working properly, want %v, got %v", want.Get(i, j), m1.Get(i, j))
		}
	}
}

func TestTransposeMatrix(t *testing.T) {
	m1 := ddd.Matrix{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m1 = m1.Transpose()

	want := ddd.Matrix{
		5, 8, 5, 2,
		1, 71, 1, 1,
		5, 2, 582, 7,
		6, 47, 4, 25,
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			if math.Abs(m1.Get(i, j)-want.Get(i, j)) < 1e-5 {
				continue
			}
			t.Fatalf("transpose matrices does not working properly, want %v, got %v", want.Get(i, j), m1.Get(i, j))
		}
	}
}
