// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

// Clamp clamps a given value in [min, max].
func Clamp(n, min, max float32) float32 {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// ClampInt clamps a given value in [min, max].
func ClampInt(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// ClampVec4 clamps a Vec4 in [min, max].
func ClampVec4(v Vec4, min, max float32) Vec4 {
	return Vec4{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
		Clamp(v.W, min, max),
	}
}

// ClampVec3 clamps a Vec4 in [min, max].
func ClampVec3(v Vec3, min, max float32) Vec3 {
	return Vec3{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
		Clamp(v.Z, min, max),
	}
}

// ClampVec2 clamps a Vec2 in [min, max].
func ClampVec2(v Vec2, min, max float32) Vec2 {
	return Vec2{
		Clamp(v.X, min, max),
		Clamp(v.Y, min, max),
	}
}
