// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import "math"

var (
	Cos        = math.Cos
	Sin        = math.Sin
	Tan        = math.Tan
	Pi         = math.Pi
	MaxInt64   = math.MaxInt64
	MaxFloat64 = math.MaxFloat64
	Round      = math.Round
	Floor      = math.Floor
	Log2       = math.Log2
	Pow        = math.Pow
	Sqrt       = math.Sqrt
)

// DefaultEpsilon is a default epsilon value for computation.
const DefaultEpsilon = 1e-7

// ApproxEq approximately compares v1 and v2.
func ApproxEq(v1, v2, epsilon float64) bool {
	return math.Abs(v1-v2) <= epsilon
}

// Clamp clamps a given value in [min, max].
func Clamp(n, min, max float64) float64 {
	return math.Min(math.Max(n, min), max)
}

// ClampV clamps a vector in [min, max].
func ClampV(v Vector, min, max float64) Vector {
	return Vector{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
		Clamp(v.W, min, max),
	}
}

// Min compares n values and returns the minimum
func Min(xs ...float64) float64 {
	min := math.MaxFloat64
	for _, x := range xs {
		min = math.Min(min, x)
	}
	return min
}

// Max compares n values and returns the maximum
func Max(xs ...float64) float64 {
	max := -math.MaxFloat32
	for _, x := range xs {
		max = math.Max(max, x)
	}
	return max
}

// ViewportMatrix returns the viewport matrix.
func ViewportMatrix(w, h float64) Matrix {
	return Matrix{
		w / 2, 0, 0, w / 2,
		0, h / 2, 0, h / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
