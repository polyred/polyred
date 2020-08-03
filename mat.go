// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

// Matrix represents a 4x4 matrix
type Matrix struct {
	x [16]float64
}

// NewMatrix creates a new identity matrix
func NewMatrix() Matrix {
	return Matrix{
		x: [16]float64{
			1, 0, 0, 0,
			0, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1,
		},
	}
}

// SetIdentity ...
func (m *Matrix) SetIdentity() *Matrix {
	m.Set(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	)
	return m
}

// Get gets the matrix elements
func (m *Matrix) Get(i, j int) float64 {
	return m.x[i+j*4]
}

// Set sets the Matrix elements
func (m *Matrix) Set(
	x11, x12, x13, x14,
	x21, x22, x23, x24,
	x31, x32, x33, x34,
	x41, x42, x43, x44 float64) {
	m.x[0] = x11
	m.x[1] = x12
	m.x[2] = x13
	m.x[3] = x14
	m.x[4] = x21
	m.x[5] = x22
	m.x[6] = x23
	m.x[7] = x24
	m.x[8] = x31
	m.x[9] = x32
	m.x[10] = x33
	m.x[11] = x34
	m.x[12] = x41
	m.x[13] = x42
	m.x[14] = x43
	m.x[15] = x44
}

// SetMat sets the matrix using given matrix
func (m *Matrix) SetMat(mm Matrix) {
	for i := 0; i < len(m.x); i++ {
		m.x[i] = mm.x[i]
	}
}

// MultiplyMatrix implements matrix multiplication for two
// 4x4 matrices and assigns the result to this.
func (m *Matrix) MultiplyMatrix(mm *Matrix) *Matrix {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m.x[i*4+k] * mm.x[k*4+j]
			}
			m.x[i*4+j] = sum
		}
	}
	return m
}

// MultiplyMatrices implements matrix multiplication for two
// 4x4 matrices and assigns the result to this.
func (m *Matrix) MultiplyMatrices(m1, m2 *Matrix) *Matrix {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			sum := 0.0
			for k := 0; k < 4; k++ {
				sum += m1.x[i*4+k] * m2.x[k*4+j]
			}
			m.x[i*4+j] = sum
		}
	}
	return m
}

// Inverse computes the inverse matrix of a given Matrix
func (m *Matrix) Inverse() *Matrix {
	inv := [16]float64{}
	inv[0] = m.x[5]*m.x[10]*m.x[15] - m.x[5]*m.x[11]*m.x[14] -
		m.x[9]*m.x[6]*m.x[15] + m.x[9]*m.x[7]*m.x[14] +
		m.x[13]*m.x[6]*m.x[11] - m.x[13]*m.x[7]*m.x[10]
	inv[4] = -m.x[4]*m.x[10]*m.x[15] + m.x[4]*m.x[11]*m.x[14] +
		m.x[8]*m.x[6]*m.x[15] - m.x[8]*m.x[7]*m.x[14] -
		m.x[12]*m.x[6]*m.x[11] + m.x[12]*m.x[7]*m.x[10]
	inv[8] = m.x[4]*m.x[9]*m.x[15] - m.x[4]*m.x[11]*m.x[13] -
		m.x[8]*m.x[5]*m.x[15] + m.x[8]*m.x[7]*m.x[13] +
		m.x[12]*m.x[5]*m.x[11] - m.x[12]*m.x[7]*m.x[9]
	inv[12] = -m.x[4]*m.x[9]*m.x[14] + m.x[4]*m.x[10]*m.x[13] +
		m.x[8]*m.x[5]*m.x[14] - m.x[8]*m.x[6]*m.x[13] -
		m.x[12]*m.x[5]*m.x[10] + m.x[12]*m.x[6]*m.x[9]
	inv[1] = -m.x[1]*m.x[10]*m.x[15] + m.x[1]*m.x[11]*m.x[14] +
		m.x[9]*m.x[2]*m.x[15] - m.x[9]*m.x[3]*m.x[14] -
		m.x[13]*m.x[2]*m.x[11] + m.x[13]*m.x[3]*m.x[10]
	inv[5] = m.x[0]*m.x[10]*m.x[15] - m.x[0]*m.x[11]*m.x[14] -
		m.x[8]*m.x[2]*m.x[15] + m.x[8]*m.x[3]*m.x[14] +
		m.x[12]*m.x[2]*m.x[11] - m.x[12]*m.x[3]*m.x[10]
	inv[9] = -m.x[0]*m.x[9]*m.x[15] + m.x[0]*m.x[11]*m.x[13] +
		m.x[8]*m.x[1]*m.x[15] - m.x[8]*m.x[3]*m.x[13] -
		m.x[12]*m.x[1]*m.x[11] + m.x[12]*m.x[3]*m.x[9]
	inv[13] = m.x[0]*m.x[9]*m.x[14] - m.x[0]*m.x[10]*m.x[13] -
		m.x[8]*m.x[1]*m.x[14] + m.x[8]*m.x[2]*m.x[13] +
		m.x[12]*m.x[1]*m.x[10] - m.x[12]*m.x[2]*m.x[9]
	inv[2] = m.x[1]*m.x[6]*m.x[15] - m.x[1]*m.x[7]*m.x[14] -
		m.x[5]*m.x[2]*m.x[15] + m.x[5]*m.x[3]*m.x[14] +
		m.x[13]*m.x[2]*m.x[7] - m.x[13]*m.x[3]*m.x[6]
	inv[6] = -m.x[0]*m.x[6]*m.x[15] + m.x[0]*m.x[7]*m.x[14] +
		m.x[4]*m.x[2]*m.x[15] - m.x[4]*m.x[3]*m.x[14] -
		m.x[12]*m.x[2]*m.x[7] + m.x[12]*m.x[3]*m.x[6]
	inv[10] = m.x[0]*m.x[5]*m.x[15] - m.x[0]*m.x[7]*m.x[13] -
		m.x[4]*m.x[1]*m.x[15] + m.x[4]*m.x[3]*m.x[13] +
		m.x[12]*m.x[1]*m.x[7] - m.x[12]*m.x[3]*m.x[5]
	inv[14] = -m.x[0]*m.x[5]*m.x[14] + m.x[0]*m.x[6]*m.x[13] +
		m.x[4]*m.x[1]*m.x[14] - m.x[4]*m.x[2]*m.x[13] -
		m.x[12]*m.x[1]*m.x[6] + m.x[12]*m.x[2]*m.x[5]
	inv[3] = -m.x[1]*m.x[6]*m.x[11] + m.x[1]*m.x[7]*m.x[10] +
		m.x[5]*m.x[2]*m.x[11] - m.x[5]*m.x[3]*m.x[10] -
		m.x[9]*m.x[2]*m.x[7] + m.x[9]*m.x[3]*m.x[6]
	inv[7] = m.x[0]*m.x[6]*m.x[11] - m.x[0]*m.x[7]*m.x[10] -
		m.x[4]*m.x[2]*m.x[11] + m.x[4]*m.x[3]*m.x[10] +
		m.x[8]*m.x[2]*m.x[7] - m.x[8]*m.x[3]*m.x[6]
	inv[11] = -m.x[0]*m.x[5]*m.x[11] + m.x[0]*m.x[7]*m.x[9] +
		m.x[4]*m.x[1]*m.x[11] - m.x[4]*m.x[3]*m.x[9] -
		m.x[8]*m.x[1]*m.x[7] + m.x[8]*m.x[3]*m.x[5]
	inv[15] = m.x[0]*m.x[5]*m.x[10] - m.x[0]*m.x[6]*m.x[9] -
		m.x[4]*m.x[1]*m.x[10] + m.x[4]*m.x[2]*m.x[9] +
		m.x[8]*m.x[1]*m.x[6] - m.x[8]*m.x[2]*m.x[5]
	det := m.x[0]*inv[0] + m.x[1]*inv[4] + m.x[2]*inv[8] + m.x[3]*inv[12]
	if det == 0 {
		panic("cannot invert matrix, det === 0")
	}
	for i := 0; i < 16; i++ {
		m.x[i] = inv[i] / det
	}
	return m
}

// Transpose computes the transpose matrix of a given Matrix
func (m *Matrix) Transpose() *Matrix {
	m.x[1], m.x[4] = m.x[4], m.x[1]
	m.x[2], m.x[8] = m.x[8], m.x[2]
	m.x[3], m.x[12] = m.x[12], m.x[3]
	m.x[6], m.x[9] = m.x[9], m.x[6]
	m.x[7], m.x[13] = m.x[13], m.x[7]
	m.x[11], m.x[14] = m.x[14], m.x[11]
	return m
}
