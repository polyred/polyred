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
func (v *Vector) Add(u *Vector) *Vector {
	v.X += u.X
	v.Y += u.Y
	v.Z += u.Z
	v.W += u.W
	return v
}

// Sub subtracts the given two vectors, or point and vector, or two points
func (v *Vector) Sub(u *Vector) *Vector {
	v.X -= u.X
	v.Y -= u.Y
	v.Z -= u.Z
	v.W -= u.W
	return v
}

// MultiplyScalar implements scalar vector or scalar point multiplication
func (v *Vector) MultiplyScalar(s float64) *Vector {
	v.X *= s
	v.Y *= s
	v.Z *= s
	v.W *= s
	return v
}

// Dot implements dot product of two vectors
func (v *Vector) Dot(u *Vector) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z + v.W*u.W
}

// CrossVectors implements cross product for two given vectors
// and assign the result to this.
func (v *Vector) CrossVectors(u1, u2 *Vector) *Vector {
	v.X = u1.Y*u2.Z - u1.Z*u2.Y
	v.Y = u1.Z*u2.X - u1.X*u2.Z
	v.Z = u1.X*u2.Y - u1.Y*u2.X
	v.W = 0
	return v
}

// Normalize normalizes this vector to a unit vector
func (v *Vector) Normalize() *Vector {
	norm := math.Sqrt(v.Dot(v))
	v.X /= norm
	v.Y /= norm
	v.Z /= norm
	v.W /= norm
	return v
}

// ApplyMatrix applies 4x4 matrix and 4x1 vector multiplication.
// the given matrix multiplies v from the left.
func (v *Vector) ApplyMatrix(m *Matrix) *Vector {
	x := m.x[0]*v.X + m.x[1]*v.Y + m.x[2]*v.Z + m.x[3]*v.W
	y := m.x[4]*v.X + m.x[5]*v.Y + m.x[6]*v.Z + m.x[7]*v.W
	z := m.x[8]*v.X + m.x[9]*v.Y + m.x[10]*v.Z + m.x[11]*v.W
	w := m.x[12]*v.X + m.x[13]*v.Y + m.x[14]*v.Z + m.x[15]*v.W

	v.X = x
	v.Y = y
	v.Z = z
	v.W = w
	return v
}
