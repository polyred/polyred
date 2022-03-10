// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

// Clamp clamps a given value in [min, max].
func Clamp[T ~int | ~int32 | ~int64 | Float](n, min, max T) T {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// ClampVec clamps a Vec4 in [min, max].
func ClampVec4[T Float](v Vec4[T], min, max T) Vec4[T] {
	return Vec4[T]{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
		Clamp(v.W, min, max),
	}
}

// ClampVec3 clamps a Vec4 in [min, max].
func ClampVec3[T Float](v Vec3[T], min, max T) Vec3[T] {
	return Vec3[T]{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
	}
}

// ClampVec2 clamps a Vec2 in [min, max].
func ClampVec2[T Float](v Vec2[T], min, max T) Vec2[T] {
	return Vec2[T]{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
	}
}
