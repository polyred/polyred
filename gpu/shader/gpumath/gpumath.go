// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gpumath is the shared math library for GPU/CPU kernels authored once
// in Go. A kernel uses these types and functions (typically via a dot-import) so
// it is ordinary Go that runs on the CPU; the Go->shader compiler (poly.red/gpu/
// shader) recognizes the same names and emits the backend shading-language
// equivalents (method calls map to operators, e.g. a.Sub(b) -> (a - b); free
// functions map to builtins, e.g. Normalize -> normalize). This lets the CPU and
// GPU renderers share one kernel source. See specs/foundations/unified-renderer.md.
//
// Vectors are float32 and column-major matrices match the backends' float4x4 /
// mat4. Operator overloading is expressed as methods because Go has none; the
// compiler lowers the methods back to operators.
package gpumath

import "math"

// Vec2, Vec3, Vec4 are float32 vectors.
type Vec2 struct{ X, Y float32 }
type Vec3 struct{ X, Y, Z float32 }
type Vec4 struct{ X, Y, Z, W float32 }

// Mat4 is a 4x4 column-major matrix (columns C0..C3), matching MSL float4x4 and
// GLSL mat4 construction order.
type Mat4 struct{ C0, C1, C2, C3 Vec4 }

// --- Vec4 methods (the compiler lowers these to operators/builtins) ---

func (a Vec4) Add(b Vec4) Vec4      { return Vec4{a.X + b.X, a.Y + b.Y, a.Z + b.Z, a.W + b.W} }
func (a Vec4) Sub(b Vec4) Vec4      { return Vec4{a.X - b.X, a.Y - b.Y, a.Z - b.Z, a.W - b.W} }
func (a Vec4) Mul(b Vec4) Vec4      { return Vec4{a.X * b.X, a.Y * b.Y, a.Z * b.Z, a.W * b.W} }
func (a Vec4) Scale(s float32) Vec4 { return Vec4{a.X * s, a.Y * s, a.Z * s, a.W * s} }
func (a Vec4) Div(s float32) Vec4   { return Vec4{a.X / s, a.Y / s, a.Z / s, a.W / s} }
func (a Vec4) Dot(b Vec4) float32   { return a.X*b.X + a.Y*b.Y + a.Z*b.Z + a.W*b.W }
func (a Vec4) Length() float32      { return float32(math.Sqrt(float64(a.Dot(a)))) }
func (a Vec4) Normalize() Vec4 {
	l := a.Length()
	if l == 0 {
		return a
	}
	return a.Scale(1 / l)
}

// --- free functions (the compiler maps these to shader builtins) ---

func Add(a, b Vec4) Vec4       { return a.Add(b) }
func Sub(a, b Vec4) Vec4       { return a.Sub(b) }
func Mul(a, b Vec4) Vec4       { return a.Mul(b) }
func Dot(a, b Vec4) float32    { return a.Dot(b) }
func Length(a Vec4) float32    { return a.Length() }
func Normalize(a Vec4) Vec4    { return a.Normalize() }
func MulV(m Mat4, v Vec4) Vec4 { return m.MulV(v) }

func Clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func Pow(x, y float32) float32 { return float32(math.Pow(float64(x), float64(y))) }
func Sqrt(x float32) float32   { return float32(math.Sqrt(float64(x))) }
func Sin(x float32) float32    { return float32(math.Sin(float64(x))) }
func Cos(x float32) float32    { return float32(math.Cos(float64(x))) }
func Tan(x float32) float32    { return float32(math.Tan(float64(x))) }
func Atan(x float32) float32   { return float32(math.Atan(float64(x))) }
func Asin(x float32) float32   { return float32(math.Asin(float64(x))) }
func Acos(x float32) float32   { return float32(math.Acos(float64(x))) }
func Exp(x float32) float32    { return float32(math.Exp(float64(x))) }
func Log(x float32) float32    { return float32(math.Log(float64(x))) }
func Floor(x float32) float32  { return float32(math.Floor(float64(x))) }
func Ceil(x float32) float32   { return float32(math.Ceil(float64(x))) }
func Round(x float32) float32  { return float32(math.Round(float64(x))) }
func Absf(x float32) float32   { return float32(math.Abs(float64(x))) }

func Minf(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func Maxf(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

// --- Mat4 ---

// MulV multiplies the matrix by a column vector (column-major: result =
// C0*v.X + C1*v.Y + C2*v.Z + C3*v.W), matching float4x4 * float4.
func (m Mat4) MulV(v Vec4) Vec4 {
	return m.C0.Scale(v.X).Add(m.C1.Scale(v.Y)).Add(m.C2.Scale(v.Z)).Add(m.C3.Scale(v.W))
}
