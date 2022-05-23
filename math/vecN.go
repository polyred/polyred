// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"fmt"
	"math/rand"
	"strings"
)

// Vec is an N dimensional vector.
type Vec[T Float] struct {
	Data []T
}

// NewVec creates a point or a vector with given parameters.
func NewVec[T Float](data ...T) Vec[T] {
	return Vec[T]{Data: append([]T{}, data...)}
}

// NewRandVec
func NewRandVec[T Float](n int) Vec[T] {
	d := make([]T, n)
	for i := range d {
		d[i] = T(rand.Float64())
	}
	return Vec[T]{Data: d}
}

// String returns a string format of the given Vec.
func (m Vec[T]) String() string {
	s := "<"
	for i := range m.Data {
		s += fmt.Sprintf("%v, ", m.Data[i])
	}
	s = strings.TrimSuffix(s, ", ") + ">"
	return s
}

// Eq checks whether two vectors are equal.
func (v Vec[T]) Eq(u Vec[T]) bool {
	for i := range v.Data {
		if !ApproxEq(v.Data[i], u.Data[i], Epsilon) {
			return false
		}
	}
	return true
}

// Add adds the given two vectors, or point and vector, or two points
func (v *Vec[T]) Add(u Vec[T]) (r Vec[T]) {
	r.Data = make([]T, len(v.Data))
	for i := range r.Data {
		r.Data[i] = v.Data[i] + u.Data[i]
	}
	return
}

// Sub subtracts the given two vectors, or point and vector, or two points
func (v Vec[T]) Sub(u Vec[T]) (r Vec[T]) {
	r.Data = make([]T, len(v.Data))
	for i := range r.Data {
		r.Data[i] = v.Data[i] - u.Data[i]
	}
	return
}

// IsZero asserts the x, y, z components of the given vector, and returns
// true if it is a zero vector or point.
func (v Vec[T]) IsZero() bool {
	for i := range v.Data {
		if !ApproxEq(v.Data[i], 0, Epsilon) {
			return false
		}
	}
	return true
}

// Scale scales the given vector using given scalars
func (v Vec[T]) Scale(w T) (r Vec[T]) {
	r.Data = make([]T, len(v.Data))
	for i := range r.Data {
		r.Data[i] = v.Data[i] * w
	}
	return
}

// Translate translates the given position or vector
func (v Vec[T]) Translate(u Vec[T]) (r Vec[T]) {
	r.Data = make([]T, len(v.Data))
	for i := range r.Data {
		r.Data[i] = v.Data[i] + u.Data[i]
	}
	return
}

// Dot implements dot product of two vectors
func (v Vec[T]) Dot(u Vec[T]) T {
	var dot T
	for i := range v.Data {
		// Use FMA to control floating number round behavior.
		// See https://go.dev/issue/52293
		dot = FMA(v.Data[i], u.Data[i], dot)
	}
	return dot
}

// Len computes the length of the given Vector
func (v Vec[T]) Len() T {
	return Sqrt(v.Dot(v))
}

// Unit normalizes this vector to an unit vector
func (v Vec[T]) Unit() (r Vec[T]) {
	return v.Scale(1.0 / v.Len())
}

// ApplyMatrix applies 4x4 matrix and 4x1 vector multiplication.
// the given matrix multiplies v from the left.
// func (v Vec[T]) Apply(m Mat4[T]) Vec[T] {
// 	x := m.X00*v.X + m.X01*v.Y + m.X02*v.Z + m.X03*v.W
// 	y := m.X10*v.X + m.X11*v.Y + m.X12*v.Z + m.X13*v.W
// 	z := m.X20*v.X + m.X21*v.Y + m.X22*v.Z + m.X23*v.W
// 	w := m.X30*v.X + m.X31*v.Y + m.X32*v.Z + m.X33*v.W
// 	return Vec[T]{x, y, z, w}
// }
