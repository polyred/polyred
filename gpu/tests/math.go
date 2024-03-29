// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests

import (
	_ "embed"

	"poly.red/math"
)

// Add is a GPU version of math.Mat[float32].Add method.
func Add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	return add(m1, m2)
}

// Sub is a GPU version of math.Mat[float32].Sub method.
func Sub[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	return sub(m1, m2)
}

// Sqrt is a GPU version of math.Mat[float32].Sub method.
func Sqrt[T DataType](m math.Mat[T]) math.Mat[T] {
	return sqrt(m)
}

// Mul is a GPU version of math.Mat[float32].Mul method.
func Mul[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	return mul(m1, m2)
}
