// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math/rand"
)

// Vec4 uses homogeneous coordinates (x, y, z, w) that represents
// either a point or a vector.
type Vec4[T Float] struct {
	X, Y, Z, W T
}

// NewVec4 creates a point or a vector with given parameters.
func NewVec4[T Float](x, y, z, w T) Vec4[T] {
	return Vec4[T]{x, y, z, w}
}

// NewRandVec4
func NewRandVec4[T Float]() Vec4[T] {
	return Vec4[T]{
		T(rand.Float64()),
		T(rand.Float64()),
		T(rand.Float64()),
		T(rand.Float64()),
	}
}

// String returns a string format of the given Vec4.
func (m Vec4[T]) String() string {
	return fmt.Sprintf(`<%v, %v, %v, %v>`, m.X, m.Y, m.Z, m.W)
}

// Eq checks whether two vectors are equal.
func (v Vec4[T]) Eq(u Vec4[T]) bool {
	return ApproxEq(v.X, u.X, Epsilon) &&
		ApproxEq(v.Y, u.Y, Epsilon) &&
		ApproxEq(v.Z, u.Z, Epsilon) &&
		ApproxEq(v.W, u.W, Epsilon)
}

// Less compares whether all components of v is less than the given u.
func (v Vec4[T]) Less(u Vec4[T]) bool {
	return ApproxLess(v.X, u.X, Epsilon) &&
		ApproxLess(v.Y, u.Z, Epsilon) &&
		ApproxLess(v.Z, u.Z, Epsilon) &&
		ApproxEq(v.W, u.W, Epsilon)
}

// Add adds the given two vectors, or point and vector, or two points
func (v *Vec4[T]) Add(u Vec4[T]) Vec4[T] {
	return Vec4[T]{v.X + u.X, v.Y + u.Y, v.Z + u.Z, v.W + u.W}
}

// Sub subtracts the given two vectors, or point and vector, or two points
func (v Vec4[T]) Sub(u Vec4[T]) Vec4[T] {
	return Vec4[T]{v.X - u.X, v.Y - u.Y, v.Z - u.Z, v.W - u.W}
}

// IsZero asserts the x, y, z components of the given vector, and returns
// true if it is a zero vector or point.
func (v Vec4[T]) IsZero() bool {
	return ApproxEq(v.X, 0, Epsilon) &&
		ApproxEq(v.Y, 0, Epsilon) &&
		ApproxEq(v.Z, 0, Epsilon)
}

// Scale scales the given vector using given scalars
func (v Vec4[T]) Scale(x, y, z, w T) Vec4[T] {
	return Vec4[T]{v.X * x, v.Y * y, v.Z * z, v.W * w}
}

// Translate translates the given position or vector
func (v Vec4[T]) Translate(x, y, z T) Vec4[T] {
	if v.W == 0 {
		return v
	}
	invW := 1 / v.W
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	return Vec4[T]{FMA(v.X, invW, x), FMA(v.Y, invW, y), FMA(v.Z, invW, z), 1}
}

// Dot implements dot product of two vectors
func (v Vec4[T]) Dot(u Vec4[T]) T {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	return FMA(v.X, u.X, FMA(v.Y, u.Y, FMA(v.Z, u.Z, v.W*u.W)))
}

// Len computes the length of the given Vector
func (v Vec4[T]) Len() T {
	return Sqrt(v.Dot(v))
}

// Unit normalizes this vector to an unit vector
func (v Vec4[T]) Unit() Vec4[T] {
	n := 1.0 / v.Len()
	return Vec4[T]{v.X * n, v.Y * n, v.Z * n, v.W * n}
}

// ApplyMatrix applies 4x4 matrix and 4x1 vector multiplication.
// the given matrix multiplies v from the left.
func (v Vec4[T]) Apply(m Mat4[T]) Vec4[T] {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	x := FMA(m.X00, v.X, FMA(m.X01, v.Y, FMA(m.X02, v.Z, m.X03*v.W)))
	y := FMA(m.X10, v.X, FMA(m.X11, v.Y, FMA(m.X12, v.Z, m.X13*v.W)))
	z := FMA(m.X20, v.X, FMA(m.X21, v.Y, FMA(m.X22, v.Z, m.X23*v.W)))
	w := FMA(m.X30, v.X, FMA(m.X31, v.Y, FMA(m.X32, v.Z, m.X33*v.W)))
	return Vec4[T]{x, y, z, w}
}

// ToVec2 drops the z and w components of the given Vec4[T].
func (v Vec4[T]) ToVec2() Vec2[T] {
	return Vec2[T]{v.X, v.Y}
}

// ToVec3 drops the w component of the given Vec4[T].
func (v Vec4[T]) ToVec3() Vec3[T] {
	return Vec3[T]{v.X, v.Y, v.Z}
}

// Cross applies cross product of two given vectors
// and returns the resulting vector.
func (v Vec4[T]) Cross(u Vec4[T]) Vec4[T] {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	x := FMA(v.Y, u.Z, -v.Z*u.Y)
	y := FMA(v.Z, u.X, -v.X*u.Z)
	z := FMA(v.X, u.Y, -v.Y*u.X)
	return Vec4[T]{x, y, z, 0}
}

// Pos converts a homogeneous represented vector to a point
func (v Vec4[T]) Pos() Vec4[T] {
	if v.W == 1 || v.W == 0 {
		return Vec4[T]{v.X, v.Y, v.Z, 1}
	}
	invW := 1.0 / v.W
	return Vec4[T]{v.X * invW, v.Y * invW, v.Z * invW, 1}
}

// Vec converts the a homogeneous represented point to a vector
func (v Vec4[T]) Vec() Vec4[T] {
	if v.W == 0 || v.W == 1 {
		return Vec4[T]{v.X, v.Y, v.Z, 0}
	}
	invW := 1.0 / v.W
	return Vec4[T]{v.X * invW, v.Y * invW, v.Z * invW, 0}
}
