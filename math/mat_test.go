// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"testing"

	"changkun.de/x/polyred/math"
)

func TestNewMatrix(t *testing.T) {
	m := math.Mat4I

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

func TestMat4_Get(t *testing.T) {
	m := math.Mat4{
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

func TestMat4_MulM(t *testing.T) {
	m1 := math.Mat4{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	m2 := math.Mat4{
		16, 15, 14, 13,
		12, 11, 10, 9,
		8, 7, 6, 5,
		4, 3, 2, 1,
	}

	got := m1.MulM(m2)

	want := math.Mat4{
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

func TestMat4_Inv(t *testing.T) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m1 = m1.Inv()

	want := math.Mat4{
		1003995.0 / 4463716, -10967.0 / 4463716, -5949.0 / 4463716, -219389.0 / 4463716,
		-62879.0 / 4463716, 65251.0 / 4463716, 1613.0 / 4463716, -107839.0 / 4463716,
		-3999.0 / 2231858, -3.0 / 2231858, 3865.0 / 2231858, 347.0 / 2231858,
		-75565.0 / 4463716, -1731.0 / 4463716, -1753.0 / 4463716, 200219.0 / 4463716,
	}

	if m1.Eq(want) {
		return
	}

	t.Fatalf("inverse matrices does not working properly, got %+v, want %+v", m1, want)
}

func TestMat4_T(t *testing.T) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m1 = m1.T()

	want := math.Mat4{
		5, 8, 5, 2,
		1, 71, 1, 1,
		5, 2, 582, 7,
		6, 47, 4, 25,
	}
	if m1.Eq(want) {
		return
	}
	t.Fatalf("transpose matrices does not working properly, got %+v, want %+v", m1, want)
}

func BenchmarkMat4_Eq(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m bool
	for i := 0; i < b.N; i++ {
		m = m1.Eq(m2)
	}
	_ = m
}

func BenchmarkMat4_Add(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = m1.Add(m2)
	}
	_ = m
}

func BenchmarkMat4_Sub(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = m1.Sub(m2)
	}
	_ = m
}

func BenchmarkMat4_MulM(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = m1.MulM(m2)
	}
	_ = m
}

func BenchmarkMat4_MulV(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Vec4{
		5, 1, 5, 6,
	}

	var m math.Vec4
	for i := 0; i < b.N; i++ {
		m = m1.MulV(m2)
	}
	_ = m
}

func BenchmarkMat4_Inv(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = m1.Inv()
	}
	_ = m
}

func BenchmarkMat4_Det(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m float64
	for i := 0; i < b.N; i++ {
		m = m1.Det()
	}
	_ = m
}

func BenchmarkMat4_T(b *testing.B) {
	m1 := math.Mat4{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = m1.T()
	}
	_ = m
}
