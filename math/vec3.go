// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math/rand"
)

// Vec3 represents a 3D vector (x, y, z).
type Vec3[T Float] struct {
	X, Y, Z T
}

// NewVec3 creates a 3D vector with given parameters.
func NewVec3[T Float](x, y, z T) Vec3[T] {
	return Vec3[T]{x, y, z}
}

// NewRandVec2 returns a random 3D vector where all components are
// sitting in range [0, 1].
func NewRandVec3[T Float]() Vec3[T] {
	return Vec3[T]{
		T(rand.Float64()),
		T(rand.Float64()),
		T(rand.Float64()),
	}
}

// String returns a string format of the given Vec3.
func (m Vec3[T]) String() string {
	return fmt.Sprintf(`<%v, %v, %v>`, m.X, m.Y, m.Z)
}

// Eq compares the two given vectors, and returns true if they are equal.
func (v Vec3[T]) Eq(u Vec3[T]) bool {
	return ApproxEq(v.X, u.X, Epsilon) &&
		ApproxEq(v.Y, u.Y, Epsilon) &&
		ApproxEq(v.Z, u.Z, Epsilon)
}

// Less compares whether all components of v is less than the given u.
func (v Vec3[T]) Less(u Vec3[T]) bool {
	return ApproxLess(v.X, u.X, Epsilon) &&
		ApproxLess(v.Y, u.Z, Epsilon) &&
		ApproxLess(v.Z, u.Z, Epsilon)
}

// Add add the two given vectors, and returns the resulting vector.
func (v Vec3[T]) Add(u Vec3[T]) Vec3[T] {
	return Vec3[T]{v.X + u.X, v.Y + u.Y, v.Z + u.Z}
}

// Sub subtracts the two given vectors, and returns the resulting vector.
func (v Vec3[T]) Sub(u Vec3[T]) Vec3[T] {
	return Vec3[T]{v.X - u.X, v.Y - u.Y, v.Z - u.Z}
}

// IsZero checks if the given vector is a zero vector.
func (v Vec3[T]) IsZero() bool {
	return ApproxEq(v.X, 0, Epsilon) && ApproxEq(v.Y, 0, Epsilon) && ApproxEq(v.Z, 0, Epsilon)
}

// Scale scales the given 3D vector and returns the resulting vector.
func (v Vec3[T]) Scale(x, y, z T) Vec3[T] {
	return Vec3[T]{v.X * x, v.Y * y, v.Z * z}
}

// Translate translates the given 3D vector and returns the resulting
// vector.
func (v Vec3[T]) Translate(x, y, z T) Vec3[T] {
	return Vec3[T]{v.X + x, v.Y + y, v.Z + z}
}

// Dot computes the dot product of two given vectors.
func (v Vec3[T]) Dot(u Vec3[T]) T {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	return FMA(v.X, u.X, FMA(v.Y, u.Y, v.Z*u.Z))
}

// Len returns the length of the given vector.
func (v Vec3[T]) Len() T {
	return Sqrt(v.Dot(v))
}

// Unit computes the unit vector along the direction of the given vector.
func (v Vec3[T]) Unit() Vec3[T] {
	n := 1.0 / v.Len()
	return Vec3[T]{v.X * n, v.Y * n, v.Z * n}
}

// ToVec4 converts to a Vec3 to Vec4 using the given w component.
func (v Vec3[T]) ToVec4(w T) Vec4[T] {
	return Vec4[T]{v.X, v.Y, v.Z, w}
}

// Apply applies the 3D matrix multiplication to the given vector on the
// left side and returns the resulting 3D vector.
func (v Vec3[T]) Apply(m Mat3[T]) Vec3[T] {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	x := FMA(m.X00, v.X, FMA(m.X01, v.Y, m.X02*v.Z))
	y := FMA(m.X10, v.X, FMA(m.X11, v.Y, m.X12*v.Z))
	z := FMA(m.X20, v.X, FMA(m.X21, v.Y, m.X22*v.Z))
	return Vec3[T]{x, y, z}
}

// Cross applies cross product of two given vectors and returns the
// resulting vector.
func (v Vec3[T]) Cross(u Vec3[T]) Vec3[T] {
	// Use FMA to control floating number round behavior.
	// See https://go.dev/issue/52293
	x := FMA(v.Y, u.Z, -v.Z*u.Y)
	y := FMA(v.Z, u.X, -v.X*u.Z)
	z := FMA(v.X, u.Y, -v.Y*u.X)
	return Vec3[T]{x, y, z}
}
