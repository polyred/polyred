// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import "fmt"

// Mat2I is an identity Mat2
func Mat2I[T Float]() Mat2[T] {
	return Mat2[T]{
		1, 0,
		0, 1,
	}
}

// Mat2Zero is a zero Mat2
func Mat2Zero[T Float]() Mat2[T] {
	return Mat2[T]{
		0, 0,
		0, 0,
	}
}

// Mat2 represents a 2x2 Mat2:
//
// / X00, X01 \
// \ X10, X11 /
type Mat2[T Float] struct {
	// This is the best implementation that benefits from compiler
	// optimization, which exports all elements of a 3x4 Mat2.
	// See benchmarks at https://golang.design/research/pointer-params/.
	X00, X01 T
	X10, X11 T
}

// NewMat2 returns a new Mat2.
func NewMat2[T Float](
	X00, X01,
	X10, X11 T) Mat2[T] {
	return Mat2[T]{
		X00, X01,
		X10, X11,
	}
}

// String returns a string format of the given matrix.
func (m Mat2[T]) String() string {
	return fmt.Sprintf(`[[%v, %v], [%v, %v]]`, m.X00, m.X01, m.X10, m.X11)
}

// Get gets the Mat2 elements
func (m Mat2[T]) Get(i, j int) T {
	if i < 0 || i > 1 || j < 0 || j > 1 {
		panic("invalid index")
	}

	switch i*2 + j {
	case 0:
		return m.X00
	case 1:
		return m.X01
	case 2:
		return m.X10
	case 3:
		fallthrough
	default:
		return m.X11
	}
}

// Set set the Mat2 elements at row i and column j
func (m *Mat2[T]) Set(i, j int, v T) {
	if i < 0 || i > 1 || j < 0 || j > 1 {
		panic("invalid index")
	}

	switch i*2 + j {
	case 0:
		m.X00 = v
	case 1:
		m.X01 = v
	case 2:
		m.X10 = v
	case 3:
		fallthrough
	default:
		m.X11 = v
	}
}

// Eq checks whether the given two matrices are equal or not.
func (m Mat2[T]) Eq(n Mat2[T]) bool {
	if ApproxEq(m.X00, n.X00, Epsilon) &&
		ApproxEq(m.X10, n.X10, Epsilon) &&
		ApproxEq(m.X01, n.X01, Epsilon) &&
		ApproxEq(m.X11, n.X11, Epsilon) {
		return true
	}
	return false
}

func (m Mat2[T]) Add(n Mat2[T]) Mat2[T] {
	return Mat2[T]{
		m.X00 + n.X00,
		m.X01 + n.X01,
		m.X10 + n.X10,
		m.X11 + n.X11,
	}
}

func (m Mat2[T]) Sub(n Mat2[T]) Mat2[T] {
	return Mat2[T]{
		m.X00 - n.X00, m.X01 - n.X01,
		m.X10 - n.X10, m.X11 - n.X11,
	}
}

// Mul implements Mat2 multiplication for two
// 3x3 matrices and assigns the result to this.
func (m Mat2[T]) MulM(n Mat2[T]) Mat2[T] {
	mm := Mat2[T]{}
	mm.X00 = m.X00*n.X00 + m.X01*n.X10
	mm.X10 = m.X10*n.X00 + m.X11*n.X10
	mm.X01 = m.X00*n.X01 + m.X01*n.X11
	mm.X11 = m.X10*n.X01 + m.X11*n.X11
	return mm
}

// MulVec implements Mat2 vector multiplication
// and returns the resulting vector.
func (m Mat2[T]) MulV(v Vec2[T]) Vec2[T] {
	x := m.X00*v.X + m.X01*v.Y
	y := m.X10*v.X + m.X11*v.Y
	return Vec2[T]{x, y}
}

// Det computes the determinant of the Mat2
func (m Mat2[T]) Det() T {
	return m.X00*m.X11 - m.X01*m.X10
}

// T computes the transpose Mat2 of a given Mat2
func (m Mat2[T]) T() Mat2[T] {
	return Mat2[T]{
		m.X00, m.X10,
		m.X01, m.X11,
	}
}
