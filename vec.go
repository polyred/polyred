// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import (
	"math"
)

// Vector uses homogeneous coordinates (x, y, z, w) that represents
// either a point or a vector.
type Vector struct {
	X, Y, Z, W float64
}

// NewVector creates a point or a vector with given parameters
func NewVector(x, y, z, w float64) Vector {
	return Vector{x, y, z, w}
}

// Add adds the given two vectors, or point and vector, or two points
func (v Vector) Add(u Vector) Vector {
	return Vector{v.X + u.X, v.Y + u.Y, v.Z + u.Z, v.W + u.W}
}

// Sub subtracts the given two vectors, or point and vector, or two points
func (v Vector) Sub(u Vector) Vector {
	return Vector{v.X - u.X, v.Y - u.Y, v.Z - u.Z, v.W + u.W}
}

// Mul implements scalar vector or scalar point multiplication
func (v Vector) Mul(s float64) Vector {
	return Vector{v.X * s, v.Y * s, v.Z * s, v.W * s}
}

// Dot implements dot product of two vectors
func (v Vector) Dot(u Vector) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z + v.W*u.W
}

// Cross implements cross product for two given vectors
// and assign the result to this.
func (v Vector) Cross(u Vector) Vector {
	x := v.Y*u.Z - v.Z*u.Y
	y := v.Z*u.X - v.X*u.Z
	z := v.X*u.Y - v.Y*u.X
	return Vector{x, y, z, 0}
}

// Normalize normalizes this vector to a unit vector
func (v Vector) Normalize() Vector {
	n := 1.0 / math.Sqrt(v.Dot(v))
	return Vector{v.X * n, v.Y * n, v.Z * n, v.W * n}
}

// ApplyMatrix applies 4x4 matrix and 4x1 vector multiplication.
// the given matrix multiplies v from the left.
func (v Vector) ApplyMatrix(m *Matrix) Vector {
	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z + m.X03*v.W
	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z + m.X13*v.W
	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z + m.X23*v.W
	w := m.X30*v.X + m.X31*v.Y + m.X32*v.Z + m.X33*v.W
	return Vector{x, y, z, w}
}
