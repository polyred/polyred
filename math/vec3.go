// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"math"
	"math/rand"
)

// Vec3 represents a 3D vector (x, y, z).
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 creates a 3D vector with given parameters.
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

// NewRandVec2 returns a random 3D vector where all components are
// sitting in range [0, 1].
func NewRandVec3() Vec3 {
	return Vec3{
		rand.Float64(),
		rand.Float64(),
		rand.Float64(),
	}
}

// Eq compares the two given vectors, and returns true if they are equal.
func (v Vec3) Eq(u Vec3) bool {
	if ApproxEq(v.X, u.X, Epsilon) &&
		ApproxEq(v.Y, u.Y, Epsilon) &&
		ApproxEq(v.Z, u.Z, Epsilon) {
		return true
	}
	return false
}

// Add add the two given vectors, and returns the resulting vector.
func (v Vec3) Add(u Vec3) Vec3 {
	return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

// Sub subtracts the two given vectors, and returns the resulting vector.
func (v Vec3) Sub(u Vec3) Vec3 {
	return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

// IsZero checks if the given vector is a zero vector.
func (v Vec3) IsZero() bool {
	if ApproxEq(v.X, 0, Epsilon) &&
		ApproxEq(v.Y, 0, Epsilon) &&
		ApproxEq(v.Z, 0, Epsilon) {
		return true
	}
	return false
}

// Scale scales the given 3D vector and returns the resulting vector.
func (v Vec3) Scale(x, y, z float64) Vec3 {
	return Vec3{v.X * x, v.Y * y, v.Z * z}
}

// Translate translates the given 3D vector and returns the resulting
// vector.
func (v Vec3) Translate(x, y, z float64) Vec3 {
	return Vec3{v.X + x, v.Y + y, v.Z + z}
}

// Dot computes the dot product of two given vectors.
func (v Vec3) Dot(u Vec3) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z
}

// Len returns the length of the given vector.
func (v Vec3) Len() float64 {
	return math.Sqrt(v.Dot(v))
}

// Unit computes the unit vector along the direction of the given vector.
func (v Vec3) Unit() Vec3 {
	n := 1.0 / v.Len()
	return Vec3{v.X * n, v.Y * n, v.Z * n}
}

// ToVec4 converts to a Vec3 to Vec4 using the given w component.
func (v Vec3) ToVec4(w float64) Vec4 {
	return Vec4{v.X, v.Y, v.Z, w}
}

// Apply applies the 3D matrix multiplication to the given vector on the
// left side and returns the resulting 3D vector.
func (v Vec3) Apply(m Mat3) Vec3 {
	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z
	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z
	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z
	return Vec3{x, y, z}
}

// Cross applies cross product of two given vectors and returns the
// resulting vector.
func (v Vec3) Cross(u Vec3) Vec3 {
	x := v.Y*u.Z - v.Z*u.Y
	y := v.Z*u.X - v.X*u.Z
	z := v.X*u.Y - v.Y*u.X
	return Vec3{x, y, z}
}
