// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math/rand"
)

// Vec2 represents a 2D vector (x, y).
type Vec2[T Float] struct {
	X, Y T
}

// NewVec2 creates a 2D vector with given parameters.
func NewVec2[T Float](x, y T) Vec2[T] {
	return Vec2[T]{x, y}
}

// NewRandVec2 returns a random 2D vector where all components are
// sitting in range [0, 1].
func NewRandVec2[T Float]() Vec2[T] {
	return Vec2[T]{
		T(rand.Float64()),
		T(rand.Float64()),
	}
}

// String returns a string format of the given Vec2.
func (m Vec2[T]) String() string {
	return fmt.Sprintf(`<%v, %v>`, m.X, m.Y)
}

// Eq compares the two given vectors, and returns true if they are equal.
func (v Vec2[T]) Eq(u Vec2[T]) bool {
	if ApproxEq(v.X, u.X, Epsilon) &&
		ApproxEq(v.Y, u.Y, Epsilon) {
		return true
	}
	return false
}

// Add add the two given vectors, and returns the resulting vector.
func (v Vec2[T]) Add(u Vec2[T]) Vec2[T] {
	return Vec2[T]{v.X + u.X, v.Y + u.Y}
}

// Sub subtracts the two given vectors, and returns the resulting vector.
func (v Vec2[T]) Sub(u Vec2[T]) Vec2[T] {
	return Vec2[T]{v.X - u.X, v.Y - u.Y}
}

// IsZero checks if the given vector is a zero vector.
func (v Vec2[T]) IsZero() bool {
	if ApproxEq(v.X, 0, Epsilon) &&
		ApproxEq(v.Y, 0, Epsilon) {
		return true
	}
	return false
}

// Scale scales the given 2D vector and returns the resulting vector.
func (v Vec2[T]) Scale(x, y T) Vec2[T] {
	return Vec2[T]{v.X * x, v.Y * y}
}

// Translate translates the given 2D vector and returns the resulting vector.
func (v Vec2[T]) Translate(x, y T) Vec2[T] {
	return Vec2[T]{v.X + x, v.Y + y}
}

// Dot computes the dot product of two given vectors.
func (v Vec2[T]) Dot(u Vec2[T]) T {
	return v.X*u.X + v.Y*u.Y
}

// Len returns the length of the given vector.
func (v Vec2[T]) Len() T {
	return T(Sqrt(v.Dot(v)))
}

// Unit computes the unit vector along the direction of the given vector.
func (v Vec2[T]) Unit() Vec2[T] {
	n := 1.0 / v.Len()
	return Vec2[T]{v.X * n, v.Y * n}
}

// Apply applies the 2D matrix multiplication to the given vector on the
// left side and returns the resulting 2D vector.
func (v Vec2[T]) Apply(m Mat2[T]) Vec2[T] {
	x := m.X00*v.X + m.X01*v.Y
	y := m.X10*v.X + m.X11*v.Y
	return Vec2[T]{x, y}
}
