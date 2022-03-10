// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"testing"

	"poly.red/math"
)

func BenchmarkMat4_Eq(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4[float32]{
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
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.Add(m2)
	}
	_ = m
}

func BenchmarkMat4_Sub(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.Sub(m2)
	}
	_ = m
}

func BenchmarkMat4_MulM(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.MulM(m2)
	}
	_ = m
}

func BenchmarkMat4_MulV(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}
	m2 := math.Vec4[float32]{
		5, 1, 5, 6,
	}

	var m math.Vec4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.MulV(m2)
	}
	_ = m
}

func BenchmarkMat4_Inv(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.Inv()
	}
	_ = m
}

func BenchmarkMat4_Det(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m float32
	for i := 0; i < b.N; i++ {
		m = m1.Det()
	}
	_ = m
}

func BenchmarkMat4_T(b *testing.B) {
	m1 := math.Mat4[float32]{
		5, 1, 5, 6,
		8, 71, 2, 47,
		5, 1, 582, 4,
		2, 1, 7, 25,
	}

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = m1.T()
	}
	_ = m
}
