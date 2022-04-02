// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import "fmt"

// Mat3I is an identity Mat3
func Mat3I[T Float]() Mat3[T] {
	return Mat3[T]{
		1, 0, 0,
		0, 1, 0,
		0, 0, 1,
	}
}

// Mat3Zero is a zero Mat3
func Mat3Zero[T Float]() Mat3[T] {
	return Mat3[T]{
		0, 0, 0,
		0, 0, 0,
		0, 0, 0,
	}
}

// Mat3 represents a 3x3 Mat3
//
// / X00, X01, X02 \
// | X10, X11, X12 |
// \ X20, X21, X22 /
type Mat3[T Float] struct {
	// This is the best implementation that benefits from compiler
	// optimization, which exports all elements of a 3x4 Mat3.
	// See benchmarks at https://golang.design/research/pointer-params/.
	X00, X01, X02 T
	X10, X11, X12 T
	X20, X21, X22 T
}

func NewMat3[T Float](
	X00, X01, X02,
	X10, X11, X12,
	X20, X21, X22 T) Mat3[T] {
	return Mat3[T]{
		X00, X01, X02,
		X10, X11, X12,
		X20, X21, X22,
	}
}

// String returns a string format of the given Mat3.
func (m Mat3[T]) String() string {
	return fmt.Sprintf(`[
	[%v, %v, %v],
	[%v, %v, %v],
	[%v, %v, %v],
]`,
		m.X00, m.X01, m.X02,
		m.X10, m.X11, m.X12,
		m.X20, m.X21, m.X22)
}

// Get gets the Mat3 elements
func (m Mat3[T]) Get(i, j int) T {
	if i < 0 || i > 2 || j < 0 || j > 2 {
		panic("invalid index")
	}

	switch i*3 + j {
	case 0:
		return m.X00
	case 1:
		return m.X01
	case 2:
		return m.X02
	case 3:
		return m.X10
	case 4:
		return m.X11
	case 5:
		return m.X12
	case 6:
		return m.X20
	case 7:
		return m.X21
	case 8:
		fallthrough
	default:
		return m.X22
	}
}

// Set set the Mat3 elements at row i and column j
func (m *Mat3[T]) Set(i, j int, v T) {
	if i < 0 || i > 2 || j < 0 || j > 2 {
		panic("invalid index")
	}

	switch i*3 + j {
	case 0:
		m.X00 = v
	case 1:
		m.X01 = v
	case 2:
		m.X02 = v
	case 3:
		m.X10 = v
	case 4:
		m.X11 = v
	case 5:
		m.X12 = v
	case 6:
		m.X20 = v
	case 7:
		m.X21 = v
	case 8:
		fallthrough
	default:
		m.X22 = v
	}
}

// Eq checks whether the given two matrices are equal or not.
func (m Mat3[T]) Eq(n Mat3[T]) bool {
	return ApproxEq(m.X00, n.X00, Epsilon) &&
		ApproxEq(m.X10, n.X10, Epsilon) &&
		ApproxEq(m.X20, n.X20, Epsilon) &&
		ApproxEq(m.X01, n.X01, Epsilon) &&
		ApproxEq(m.X11, n.X11, Epsilon) &&
		ApproxEq(m.X21, n.X21, Epsilon) &&
		ApproxEq(m.X02, n.X02, Epsilon) &&
		ApproxEq(m.X12, n.X12, Epsilon) &&
		ApproxEq(m.X22, n.X22, Epsilon)
}

func (m Mat3[T]) Add(n Mat3[T]) Mat3[T] {
	return Mat3[T]{
		m.X00 + n.X00,
		m.X01 + n.X01,
		m.X02 + n.X02,
		m.X10 + n.X10,
		m.X11 + n.X11,
		m.X12 + n.X12,
		m.X20 + n.X20,
		m.X21 + n.X21,
		m.X22 + n.X22,
	}
}

func (m Mat3[T]) Sub(n Mat3[T]) Mat3[T] {
	return Mat3[T]{
		m.X00 - n.X00, m.X01 - n.X01, m.X02 - n.X02,
		m.X10 - n.X10, m.X11 - n.X11, m.X12 - n.X12,
		m.X20 - n.X20, m.X21 - n.X21, m.X22 - n.X22,
	}
}

// Mul implements Mat3 multiplication for two
// 3x3 matrices and assigns the result to this.
func (m Mat3[T]) MulM(n Mat3[T]) Mat3[T] {
	mm := Mat3[T]{}
	mm.X00 = m.X00*n.X00 + m.X01*n.X10 + m.X02*n.X20
	mm.X10 = m.X10*n.X00 + m.X11*n.X10 + m.X12*n.X20
	mm.X20 = m.X20*n.X00 + m.X21*n.X10 + m.X22*n.X20
	mm.X01 = m.X00*n.X01 + m.X01*n.X11 + m.X02*n.X21
	mm.X11 = m.X10*n.X01 + m.X11*n.X11 + m.X12*n.X21
	mm.X21 = m.X20*n.X01 + m.X21*n.X11 + m.X22*n.X21
	mm.X02 = m.X00*n.X02 + m.X01*n.X12 + m.X02*n.X22
	mm.X12 = m.X10*n.X02 + m.X11*n.X12 + m.X12*n.X22
	mm.X22 = m.X20*n.X02 + m.X21*n.X12 + m.X22*n.X22
	return mm
}

// MulVec implements Mat3[T] vector multiplication
// and returns the resulting vector.
func (m Mat3[T]) MulV(v Vec3[T]) Vec3[T] {
	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z
	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z
	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z
	return Vec3[T]{x, y, z}
}

// Det computes the determinant of the given matrix.
func (m Mat3[T]) Det() T {
	return (m.X00*m.X11*m.X22 - m.X21*m.X12) -
		m.X10*(m.X01*m.X22-m.X21*m.X02) +
		m.X20*(m.X01*m.X12-m.X11*m.X02)
}

// T computes the transpose Mat3[T] of a given Mat3[T]
func (m Mat3[T]) T() Mat3[T] {
	return Mat3[T]{
		m.X00, m.X10, m.X20,
		m.X01, m.X11, m.X21,
		m.X02, m.X12, m.X22,
	}
}
