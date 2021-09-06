// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math"
	"math/rand"
)

// Vec4 uses homogeneous coordinates (x, y, z, w) that represents
// either a point or a vector.
type Vec4 struct {
	X, Y, Z, W float64
}

// NewVec4 creates a point or a vector with given parameters.
func NewVec4(x, y, z, w float64) Vec4 {
	return Vec4{x, y, z, w}
}

// NewRandVec4
func NewRandVec4() Vec4 {
	return Vec4{
		rand.Float64(),
		rand.Float64(),
		rand.Float64(),
		rand.Float64(),
	}
}

// String returns a string format of the given Vec4.
func (m Vec4) String() string {
	return fmt.Sprintf(`<%v, %v, %v, %v>`, m.X, m.Y, m.Z, m.W)
}

// Eq checks whether two vectors are equal.
func (v Vec4) Eq(u Vec4) bool {
	if ApproxEq(v.X, u.X, Epsilon) &&
		ApproxEq(v.Y, u.Y, Epsilon) &&
		ApproxEq(v.Z, u.Z, Epsilon) &&
		ApproxEq(v.W, u.W, Epsilon) {
		return true
	}
	return false
}

// Add adds the given two vectors, or point and vector, or two points
func (v *Vec4) Add(u Vec4) Vec4 {
	return Vec4{v.X + u.X, v.Y + u.Y, v.Z + u.Z, v.W + u.W}
}

// Sub subtracts the given two vectors, or point and vector, or two points
func (v Vec4) Sub(u Vec4) Vec4 {
	return Vec4{v.X - u.X, v.Y - u.Y, v.Z - u.Z, v.W - u.W}
}

// IsZero asserts the x, y, z components of the given vector, and returns
// true if it is a zero vector or point.
func (v Vec4) IsZero() bool {
	if ApproxEq(v.X, 0, Epsilon) &&
		ApproxEq(v.Y, 0, Epsilon) &&
		ApproxEq(v.Z, 0, Epsilon) {
		return true
	}
	return false
}

// Scale scales the given vector using given scalars
func (v Vec4) Scale(x, y, z, w float64) Vec4 {
	return Vec4{v.X * x, v.Y * y, v.Z * z, v.W * w}
}

// Translate translates the given position or vector
func (v Vec4) Translate(x, y, z float64) Vec4 {
	if v.W == 0 {
		return v
	}
	return Vec4{v.X/v.W + x, v.Y/v.W + y, v.Z/v.W + z, 1}
}

// Dot implements dot product of two vectors
func (v Vec4) Dot(u Vec4) float64 {
	return v.X*u.X + v.Y*u.Y + v.Z*u.Z + v.W*u.W
}

// Len computes the length of the given Vector
func (v Vec4) Len() float64 {
	return math.Sqrt(v.Dot(v))
}

// Unit normalizes this vector to an unit vector
func (v Vec4) Unit() Vec4 {
	n := 1.0 / v.Len()
	return Vec4{v.X * n, v.Y * n, v.Z * n, v.W * n}
}

// ApplyMatrix applies 4x4 matrix and 4x1 vector multiplication.
// the given matrix multiplies v from the left.
func (v Vec4) Apply(m Mat4) Vec4 {
	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z + m.X03*v.W
	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z + m.X13*v.W
	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z + m.X23*v.W
	w := m.X30*v.X + m.X31*v.Y + m.X32*v.Z + m.X33*v.W
	return Vec4{x, y, z, w}
}

// ToVec2 drops the z and w components of the given vec4.
func (v Vec4) ToVec2() Vec2 {
	return Vec2{v.X, v.Y}
}

// ToVec3 drops the w component of the given vec4.
func (v Vec4) ToVec3() Vec3 {
	return Vec3{v.X, v.Y, v.Z}
}

// Cross applies cross product of two given vectors
// and returns the resulting vector.
func (v Vec4) Cross(u Vec4) Vec4 {
	x := v.Y*u.Z - v.Z*u.Y
	y := v.Z*u.X - v.X*u.Z
	z := v.X*u.Y - v.Y*u.X
	return Vec4{x, y, z, 0}
}

// Pos converts a homogeneous represented vector to a point
func (v Vec4) Pos() Vec4 {
	if v.W == 1 || v.W == 0 {
		return Vec4{v.X, v.Y, v.Z, 1}
	}
	return Vec4{v.X / v.W, v.Y / v.W, v.Z / v.W, 1}
}

// Vec converts the a homogeneous represented point to a vector
func (v Vec4) Vec() Vec4 {
	if v.W == 0 || v.W == 1 {
		return Vec4{v.X, v.Y, v.Z, 0}
	}
	return Vec4{v.X / v.W, v.Y / v.W, v.Z / v.W, 0}
}
