// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package math implements basic math functions which operate
// directly on float64 numbers without casting and contains
// types of common entities used in 3D Graphics such as vectors,
// matrices, quaternions and others.
package math

import "math"

// Equivalents to the standard math package.
var (
	Cos        = math.Cos
	Sin        = math.Sin
	Tan        = math.Tan
	Abs        = math.Abs
	Acos       = math.Acos
	Atan       = math.Atan
	Atan2      = math.Atan2
	Pi         = math.Pi
	Inf        = math.Inf
	MaxInt64   = math.MaxInt64
	MaxFloat64 = math.MaxFloat64
	Round      = math.Round
	Floor      = math.Floor
	Log2       = math.Log2
	Pow        = math.Pow
	Sqrt       = math.Sqrt
	IsNaN      = math.IsNaN
	Modf       = math.Modf
)

const (
	// Epsilon is a default epsilon value for computation.
	Epsilon     = 1e-7
	degToRadFac = math.Pi / 180
	radToDegFac = 180.0 / math.Pi
)

// ApproxEq approximately compares v1 and v2.
func ApproxEq(v1, v2, epsilon float64) bool {
	return math.Abs(v1-v2) <= epsilon
}

// DegToRad converts a number from degrees to radians
func DegToRad(deg float64) float64 {
	return deg * degToRadFac
}

// RadToDeg converts a number from radians to degrees
func RadToDeg(rad float64) float64 {
	return rad * radToDegFac
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
func ViewportMatrix(w, h float64) Mat4 {
	return Mat4{
		w / 2, 0, 0, w / 2,
		0, h / 2, 0, h / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
