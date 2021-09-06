// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package math implements basic math functions which operate
// directly on float32 numbers without casting and contains
// types of common entities used in 3D Graphics such as vectors,
// matrices, quaternions and others.
package math

import "math"

// Equivalents to the standard math package.
var (
	Pi         = float32(3.14159265358979323846264338327950288419716939937510582097494459)
	HalfPi     = Pi * 0.5
	TwoPi      = Pi * 2
	MaxInt32   = math.MaxInt32
	MaxInt64   = math.MaxInt64
	MaxFloat32 = float32(math.MaxFloat32)
	MaxFloat64 = float64(math.MaxFloat64)
	IsNaN      = math.IsNaN
)

const (
	// Epsilon is a default epsilon value for computation.
	Epsilon     = 1e-7
	degToRadFac = math.Pi / 180
	radToDegFac = 180.0 / math.Pi
)

// ApproxEq approximately compares v1 and v2.
func ApproxEq(v1, v2, epsilon float32) bool {
	return Abs(v1-v2) <= epsilon
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round(x float32) float32 {
	return float32(math.Round(float64(x)))
}

// Inf returns positive infinity if sign >= 0, negative infinity if sign < 0.
func Inf(sign int) float32 {
	return float32(math.Inf(sign))
}

// Pow returns x**y, the base-x exponential of y.
//
// Special cases are (in order):
//	Pow(x, ±0) = 1 for any x
//	Pow(1, y) = 1 for any y
//	Pow(x, 1) = x for any x
//	Pow(NaN, y) = NaN
//	Pow(x, NaN) = NaN
//	Pow(±0, y) = ±Inf for y an odd integer < 0
//	Pow(±0, -Inf) = +Inf
//	Pow(±0, +Inf) = +0
//	Pow(±0, y) = +Inf for finite y < 0 and not an odd integer
//	Pow(±0, y) = ±0 for y an odd integer > 0
//	Pow(±0, y) = +0 for finite y > 0 and not an odd integer
//	Pow(-1, ±Inf) = 1
//	Pow(x, +Inf) = +Inf for |x| > 1
//	Pow(x, -Inf) = +0 for |x| > 1
//	Pow(x, +Inf) = +0 for |x| < 1
//	Pow(x, -Inf) = +Inf for |x| < 1
//	Pow(+Inf, y) = +Inf for y > 0
//	Pow(+Inf, y) = +0 for y < 0
//	Pow(-Inf, y) = Pow(-0, -y)
//	Pow(x, y) = NaN for finite x < 0 and finite non-integer y
func Pow(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func Floor(x float32) float32 {
	return float32(math.Floor(float64(x)))
}

// Modf returns integer and fractional floating-point numbers
// that sum to f. Both values have the same sign as f.
//
// Special cases are:
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func Modf(f float32) (float32, float32) {
	in, fl := math.Modf(float64(f))
	return float32(in), float32(fl)
}

// Log2 returns the binary logarithm of x.
// The special cases are the same as for Log.
func Log2(x float32) float32 {
	return float32(math.Log2(float64(x)))
}

// Abs returns the absolute value of x.
//
// Special cases are:
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func Abs(x float32) float32 {
	return float32(math.Abs(float64(x)))
}

// Ceil returns the least integer value greater than or equal to x.
//
// Special cases are:
//	Ceil(±0) = ±0
//	Ceil(±Inf) = ±Inf
//	Ceil(NaN) = NaN
func Ceil(x float32) float32 {
	return float32(math.Ceil(float64(x)))
}

// Cos returns the cosine of the radian argument x.
//
// Special cases are:
//	Cos(±Inf) = NaN
//	Cos(NaN) = NaN
func Cos(x float32) float32 {
	return float32(math.Cos(float64(x)))
}

// Acos returns the arccosine, in radians, of x.
//
// Special case is:
//	Acos(x) = NaN if x < -1 or x > 1
func Acos(x float32) float32 {
	return float32(math.Acos(float64(x)))
}

// Sin returns the sine of the radian argument x.
//
// Special cases are:
//	Sin(±0) = ±0
//	Sin(±Inf) = NaN
//	Sin(NaN) = NaN
func Sin(x float32) float32 {
	return float32(math.Sin(float64(x)))
}

// Tan returns the tangent of the radian argument x.
//
// Special cases are:
//	Tan(±0) = ±0
//	Tan(±Inf) = NaN
//	Tan(NaN) = NaN
func Tan(x float32) float32 {
	return float32(math.Tan(float64(x)))
}

// Atan returns the arctangent, in radians, of x.
//
// Special cases are:
//      Atan(±0) = ±0
//      Atan(±Inf) = ±Pi/2
func Atan(x float32) float32 {
	return float32(math.Atan(float64(x)))
}

// Atan2 returns the arc tangent of y/x, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//	Atan2(y, NaN) = NaN
//	Atan2(NaN, x) = NaN
//	Atan2(+0, x>=0) = +0
//	Atan2(-0, x>=0) = -0
//	Atan2(+0, x<=-0) = +Pi
//	Atan2(-0, x<=-0) = -Pi
//	Atan2(y>0, 0) = +Pi/2
//	Atan2(y<0, 0) = -Pi/2
//	Atan2(+Inf, +Inf) = +Pi/4
//	Atan2(-Inf, +Inf) = -Pi/4
//	Atan2(+Inf, -Inf) = 3Pi/4
//	Atan2(-Inf, -Inf) = -3Pi/4
//	Atan2(y, +Inf) = 0
//	Atan2(y>0, -Inf) = +Pi
//	Atan2(y<0, -Inf) = -Pi
//	Atan2(+Inf, x) = +Pi/2
//	Atan2(-Inf, x) = -Pi/2
func Atan2(y, x float32) float32 {
	return float32(math.Atan2(float64(y), float64(x)))
}

// Sqrt returns the square root of x.
//
// Special cases are:
//	Sqrt(+Inf) = +Inf
//	Sqrt(±0) = ±0
//	Sqrt(x < 0) = NaN
//	Sqrt(NaN) = NaN
func Sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// DegToRad converts a number from degrees to radians
func DegToRad(deg float32) float32 {
	return deg * degToRadFac
}

// RadToDeg converts a number from radians to degrees
func RadToDeg(rad float32) float32 {
	return rad * radToDegFac
}

// Min compares n values and returns the minimum
func Min(xs ...float32) float32 {
	min := math.MaxFloat32
	for _, x := range xs {
		min = math.Min(float64(min), float64(x))
	}
	return float32(min)
}

// Max compares n values and returns the maximum
func Max(xs ...float32) float32 {
	max := -math.MaxFloat32
	for _, x := range xs {
		max = math.Max(float64(max), float64(x))
	}
	return float32(max)
}

// ViewportMatrix returns the viewport matrix.
func ViewportMatrix(w, h float32) Mat4 {
	return Mat4{
		w / 2, 0, 0, w / 2,
		0, h / 2, 0, h / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
