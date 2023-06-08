// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package math implements basic math functions which operate
// directly on float32 numbers without casting and contains
// types of common entities used in 3D Graphics such as vectors,
// matrices, quaternions and others.
package math

import "math"

// Equivalents to the standard math package.
const (
	Pi         = 3.14159265358979323846264338327950288419716939937510582097494459
	HalfPi     = Pi * 0.5
	TwoPi      = Pi * 2
	MaxInt32   = math.MaxInt32
	MaxInt64   = math.MaxInt64
	MaxFloat32 = math.MaxFloat32
	MaxFloat64 = math.MaxFloat64
	MaxUint32  = math.MaxUint32
)

var (
	IsNaN = math.IsNaN
)

const (
	// Epsilon is a default epsilon value for computation.
	Epsilon     = 1e-7
	degToRadFac = math.Pi / 180
	radToDegFac = 180.0 / math.Pi
)

// Float is a constraint that permits any floating-point type.
type Float interface {
	~float32 | ~float64
}

// ApproxEq compares v1 and v2 approximately.
func ApproxEq[T Float](v1, v2, epsilon T) bool {
	return Abs(v1-v2) <= epsilon
}

// ApproxLess compares whether v1 is less than v2 (v1 < v2) approximately.
func ApproxLess[T Float](v1, v2, epsilon T) bool {
	return v1 < v2 && Abs(v1-v2) > epsilon
}

// Round returns the nearest integer, rounding half away from zero.
//
// Special cases are:
//
//	Round(±0) = ±0
//	Round(±Inf) = ±Inf
//	Round(NaN) = NaN
func Round[T Float](x T) T {
	return T(math.Round(float64(x)))
}

// Inf returns positive infinity if sign >= 0, negative infinity if sign < 0.
func Inf(sign int) float32 {
	return float32(math.Inf(sign))
}

// Pow returns x**y, the base-x exponential of y.
//
// Special cases are (in order):
//
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
func Pow[T Float](x, y T) T {
	return T(math.Pow(float64(x), float64(y)))
}

// Floor returns the greatest integer value less than or equal to x.
//
// Special cases are:
//
//	Floor(±0) = ±0
//	Floor(±Inf) = ±Inf
//	Floor(NaN) = NaN
func Floor[T Float](x T) T {
	return T(math.Floor(float64(x)))
}

// Modf returns integer and fractional floating-point numbers
// that sum to f. Both values have the same sign as f.
//
// Special cases are:
//
//	Modf(±Inf) = ±Inf, NaN
//	Modf(NaN) = NaN, NaN
func Modf[T Float](f T) (T, T) {
	in, fl := math.Modf(float64(f))
	return T(in), T(fl)
}

// Log2 returns the binary logarithm of x.
// The special cases are the same as for Log.
func Log2[T Float](x T) T {
	return T(math.Log2(float64(x)))
}

// Abs returns the absolute value of x.
//
// Special cases are:
//
//	Abs(±Inf) = +Inf
//	Abs(NaN) = NaN
func Abs[T Float](x T) T {
	return T(math.Abs(float64(x)))
}

// Ceil returns the least integer value greater than or equal to x.
//
// Special cases are:
//
//	Ceil(±0) = ±0
//	Ceil(±Inf) = ±Inf
//	Ceil(NaN) = NaN
func Ceil[T Float](x T) T {
	return T(math.Ceil(float64(x)))
}

// Cos returns the cosine of the radian argument x.
//
// Special cases are:
//
//	Cos(±Inf) = NaN
//	Cos(NaN) = NaN
func Cos[T Float](x T) T {
	return T(math.Cos(float64(x)))
}

// Acos returns the arccosine, in radians, of x.
//
// Special case is:
//
//	Acos(x) = NaN if x < -1 or x > 1
func Acos[T Float](x T) T {
	return T(math.Acos(float64(x)))
}

// Sin returns the sine of the radian argument x.
//
// Special cases are:
//
//	Sin(±0) = ±0
//	Sin(±Inf) = NaN
//	Sin(NaN) = NaN
func Sin[T Float](x T) T {
	return T(math.Sin(float64(x)))
}

// Tan returns the tangent of the radian argument x.
//
// Special cases are:
//
//	Tan(±0) = ±0
//	Tan(±Inf) = NaN
//	Tan(NaN) = NaN
func Tan[T Float](x T) T {
	return T(math.Tan(float64(x)))
}

// Atan returns the arctangent, in radians, of x.
//
// Special cases are:
//
//	Atan(±0) = ±0
//	Atan(±Inf) = ±Pi/2
func Atan[T Float](x T) T {
	return T(math.Atan(float64(x)))
}

// Atan2 returns the arc tangent of y/x, using
// the signs of the two to determine the quadrant
// of the return value.
//
// Special cases are (in order):
//
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
func Atan2[T Float](y, x T) T {
	return T(math.Atan2(float64(y), float64(x)))
}

// Sqrt returns the square root of x.
//
// Special cases are:
//
//	Sqrt(+Inf) = +Inf
//	Sqrt(±0) = ±0
//	Sqrt(x < 0) = NaN
//	Sqrt(NaN) = NaN
func Sqrt[T Float](x T) T {
	return T(math.Sqrt(float64(x)))
}

// DegToRad converts a number from degrees to radians
func DegToRad[T Float](deg T) T {
	return deg * degToRadFac
}

// RadToDeg converts a number from radians to degrees
func RadToDeg[T Float](rad T) T {
	return rad * radToDegFac
}

// Min compares n values and returns the minimum
func Min[T Float](xs ...T) T {
	min := math.MaxFloat64
	for _, x := range xs {
		min = math.Min(float64(min), float64(x))
	}
	return T(min)
}

// Max compares n values and returns the maximum
func Max[T Float](xs ...T) T {
	max := -math.MaxFloat64
	for _, x := range xs {
		max = math.Max(float64(max), float64(x))
	}
	return T(max)
}

// FMA returns x * y + z, computed with only one rounding.
// (That is, FMA returns the fused multiply-add of x, y, and z.)
func FMA[T Float](x, y, z T) T {
	return T(math.FMA(float64(x), float64(y), float64(z)))
}

// ViewportMatrix returns the viewport matrix.
func ViewportMatrix[T Float](w, h T) Mat4[T] {
	return Mat4[T]{
		w / 2, 0, 0, w / 2,
		0, h / 2, 0, h / 2,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
}
