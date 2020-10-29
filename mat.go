// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

var (
	// IdentityMatrix is an identity matrix
	IdentityMatrix = Matrix{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	// ZeroMatrix is a zero matrix
	ZeroMatrix = Matrix{
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
		0, 0, 0, 0,
	}
)

// Matrix represents a 4x4 matrix
type Matrix struct {
	// This is the best implementation that benefits from compiler
	// optimization, which exports all elements of a 4x4 Matrix.
	// See benchmarks at https://golang.design/s/research.

	X00, X01, X02, X03 float64
	X10, X11, X12, X13 float64
	X20, X21, X22, X23 float64
	X30, X31, X32, X33 float64
}

// Get gets the matrix elements
func (m Matrix) Get(i, j int) float64 {
	switch i + j*4 {
	case 0:
		return m.X00
	case 1:
		return m.X01
	case 2:
		return m.X02
	case 3:
		return m.X03
	case 4:
		return m.X10
	case 5:
		return m.X11
	case 6:
		return m.X12
	case 7:
		return m.X13
	case 8:
		return m.X20
	case 9:
		return m.X21
	case 10:
		return m.X22
	case 11:
		return m.X23
	case 12:
		return m.X30
	case 13:
		return m.X31
	case 14:
		return m.X32
	default:
		return m.X33
	}
}

// Mul implements matrix multiplication for two
// 4x4 matrices and assigns the result to this.
func (m Matrix) Mul(n Matrix) Matrix {
	mm := Matrix{}
	mm.X00 = m.X00*n.X00 + m.X01*n.X10 + m.X02*n.X20 + m.X03*n.X30
	mm.X10 = m.X10*n.X00 + m.X11*n.X10 + m.X12*n.X20 + m.X13*n.X30
	mm.X20 = m.X20*n.X00 + m.X21*n.X10 + m.X22*n.X20 + m.X23*n.X30
	mm.X30 = m.X30*n.X00 + m.X31*n.X10 + m.X32*n.X20 + m.X33*n.X30
	mm.X01 = m.X00*n.X01 + m.X01*n.X11 + m.X02*n.X21 + m.X03*n.X31
	mm.X11 = m.X10*n.X01 + m.X11*n.X11 + m.X12*n.X21 + m.X13*n.X31
	mm.X21 = m.X20*n.X01 + m.X21*n.X11 + m.X22*n.X21 + m.X23*n.X31
	mm.X31 = m.X30*n.X01 + m.X31*n.X11 + m.X32*n.X21 + m.X33*n.X31
	mm.X02 = m.X00*n.X02 + m.X01*n.X12 + m.X02*n.X22 + m.X03*n.X32
	mm.X12 = m.X10*n.X02 + m.X11*n.X12 + m.X12*n.X22 + m.X13*n.X32
	mm.X22 = m.X20*n.X02 + m.X21*n.X12 + m.X22*n.X22 + m.X23*n.X32
	mm.X32 = m.X30*n.X02 + m.X31*n.X12 + m.X32*n.X22 + m.X33*n.X32
	mm.X03 = m.X00*n.X03 + m.X01*n.X13 + m.X02*n.X23 + m.X03*n.X33
	mm.X13 = m.X10*n.X03 + m.X11*n.X13 + m.X12*n.X23 + m.X13*n.X33
	mm.X23 = m.X20*n.X03 + m.X21*n.X13 + m.X22*n.X23 + m.X23*n.X33
	mm.X33 = m.X30*n.X03 + m.X31*n.X13 + m.X32*n.X23 + m.X33*n.X33
	return mm
}

// MulVec implements matrix vector multiplication
// and returns the resulting vector.
func (m Matrix) MulVec(v Vector) Vector {
	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z + m.X03*v.W
	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z + m.X13*v.W
	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z + m.X23*v.W
	w := m.X30*v.X + m.X31*v.Y + m.X32*v.Z + m.X33*v.W
	return Vector{x, y, z, w}
}

// Inverse computes the inverse matrix of a given Matrix
func (m Matrix) Inverse() Matrix {
	d := m.Det()
	if d == 0 {
		panic("zero determinant")
	}
	n := Matrix{}
	n.X00 = (m.X12*m.X23*m.X31 - m.X13*m.X22*m.X31 + m.X13*m.X21*m.X32 - m.X11*m.X23*m.X32 - m.X12*m.X21*m.X33 + m.X11*m.X22*m.X33) / d
	n.X01 = (m.X03*m.X22*m.X31 - m.X02*m.X23*m.X31 - m.X03*m.X21*m.X32 + m.X01*m.X23*m.X32 + m.X02*m.X21*m.X33 - m.X01*m.X22*m.X33) / d
	n.X02 = (m.X02*m.X13*m.X31 - m.X03*m.X12*m.X31 + m.X03*m.X11*m.X32 - m.X01*m.X13*m.X32 - m.X02*m.X11*m.X33 + m.X01*m.X12*m.X33) / d
	n.X03 = (m.X03*m.X12*m.X21 - m.X02*m.X13*m.X21 - m.X03*m.X11*m.X22 + m.X01*m.X13*m.X22 + m.X02*m.X11*m.X23 - m.X01*m.X12*m.X23) / d
	n.X10 = (m.X13*m.X22*m.X30 - m.X12*m.X23*m.X30 - m.X13*m.X20*m.X32 + m.X10*m.X23*m.X32 + m.X12*m.X20*m.X33 - m.X10*m.X22*m.X33) / d
	n.X11 = (m.X02*m.X23*m.X30 - m.X03*m.X22*m.X30 + m.X03*m.X20*m.X32 - m.X00*m.X23*m.X32 - m.X02*m.X20*m.X33 + m.X00*m.X22*m.X33) / d
	n.X12 = (m.X03*m.X12*m.X30 - m.X02*m.X13*m.X30 - m.X03*m.X10*m.X32 + m.X00*m.X13*m.X32 + m.X02*m.X10*m.X33 - m.X00*m.X12*m.X33) / d
	n.X13 = (m.X02*m.X13*m.X20 - m.X03*m.X12*m.X20 + m.X03*m.X10*m.X22 - m.X00*m.X13*m.X22 - m.X02*m.X10*m.X23 + m.X00*m.X12*m.X23) / d
	n.X20 = (m.X11*m.X23*m.X30 - m.X13*m.X21*m.X30 + m.X13*m.X20*m.X31 - m.X10*m.X23*m.X31 - m.X11*m.X20*m.X33 + m.X10*m.X21*m.X33) / d
	n.X21 = (m.X03*m.X21*m.X30 - m.X01*m.X23*m.X30 - m.X03*m.X20*m.X31 + m.X00*m.X23*m.X31 + m.X01*m.X20*m.X33 - m.X00*m.X21*m.X33) / d
	n.X22 = (m.X01*m.X13*m.X30 - m.X03*m.X11*m.X30 + m.X03*m.X10*m.X31 - m.X00*m.X13*m.X31 - m.X01*m.X10*m.X33 + m.X00*m.X11*m.X33) / d
	n.X23 = (m.X03*m.X11*m.X20 - m.X01*m.X13*m.X20 - m.X03*m.X10*m.X21 + m.X00*m.X13*m.X21 + m.X01*m.X10*m.X23 - m.X00*m.X11*m.X23) / d
	n.X30 = (m.X12*m.X21*m.X30 - m.X11*m.X22*m.X30 - m.X12*m.X20*m.X31 + m.X10*m.X22*m.X31 + m.X11*m.X20*m.X32 - m.X10*m.X21*m.X32) / d
	n.X31 = (m.X01*m.X22*m.X30 - m.X02*m.X21*m.X30 + m.X02*m.X20*m.X31 - m.X00*m.X22*m.X31 - m.X01*m.X20*m.X32 + m.X00*m.X21*m.X32) / d
	n.X32 = (m.X02*m.X11*m.X30 - m.X01*m.X12*m.X30 - m.X02*m.X10*m.X31 + m.X00*m.X12*m.X31 + m.X01*m.X10*m.X32 - m.X00*m.X11*m.X32) / d
	n.X33 = (m.X01*m.X12*m.X20 - m.X02*m.X11*m.X20 + m.X02*m.X10*m.X21 - m.X00*m.X12*m.X21 - m.X01*m.X10*m.X22 + m.X00*m.X11*m.X22) / d
	return n
}

// Det computes the determinant of the matrix
func (m Matrix) Det() float64 {
	return m.X00*m.X11*m.X22*m.X33 - m.X00*m.X11*m.X23*m.X32 +
		m.X00*m.X12*m.X23*m.X31 - m.X00*m.X12*m.X21*m.X33 +
		m.X00*m.X13*m.X21*m.X32 - m.X00*m.X13*m.X22*m.X31 -
		m.X01*m.X12*m.X23*m.X30 + m.X01*m.X12*m.X20*m.X33 -
		m.X01*m.X13*m.X20*m.X32 + m.X01*m.X13*m.X22*m.X30 -
		m.X01*m.X10*m.X22*m.X33 + m.X01*m.X10*m.X23*m.X32 +
		m.X02*m.X13*m.X20*m.X31 - m.X02*m.X13*m.X21*m.X30 +
		m.X02*m.X10*m.X21*m.X33 - m.X02*m.X10*m.X23*m.X31 +
		m.X02*m.X11*m.X23*m.X30 - m.X02*m.X11*m.X20*m.X33 -
		m.X03*m.X10*m.X21*m.X32 + m.X03*m.X10*m.X22*m.X31 -
		m.X03*m.X11*m.X22*m.X30 + m.X03*m.X11*m.X20*m.X32 -
		m.X03*m.X12*m.X20*m.X31 + m.X03*m.X12*m.X21*m.X30
}

// Transpose computes the transpose matrix of a given Matrix
func (m Matrix) Transpose() Matrix {
	return Matrix{
		m.X00, m.X10, m.X20, m.X30,
		m.X01, m.X11, m.X21, m.X31,
		m.X02, m.X12, m.X22, m.X32,
		m.X03, m.X13, m.X23, m.X33,
	}
}
