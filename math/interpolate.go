// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"image/color"
)

// Lerp computes a linear interpolation between two given numbers
// regarding the given t parameter.
func Lerp(from, to, t float32) float32 {
	return from + t*(to-from)
}

// Lerp computes a linear interpolation between two given integers
// regarding the given t parameter.
func LerpInt(from, to int, t float32) int {
	return int(float32(from) + t*float32(to-from))
}

// LerpV computes a linear interpolation between two given vectors
// regarding the given t parameter.
func LerpVec2(from, to Vec2, t float32) Vec2 {
	return Vec2{
		Lerp(from.X, to.X, t),
		Lerp(from.Y, to.Y, t),
	}
}

// LerpV computes a linear interpolation between two given vectors
// regarding the given t parameter.
func LerpVec3(from, to Vec3, t float32) Vec3 {
	return Vec3{
		Lerp(from.X, to.X, t),
		Lerp(from.Y, to.Y, t),
		Lerp(from.Z, to.Z, t),
	}
}

// LerpV computes a linear interpolation between two given vectors
// regarding the given t parameter.
func LerpVec4(from, to Vec4, t float32) Vec4 {
	return Vec4{
		Lerp(from.X, to.X, t),
		Lerp(from.Y, to.Y, t),
		Lerp(from.Z, to.Z, t),
		Lerp(from.W, to.W, t),
	}
}

// LerpC computes a linear interpolation between two given colors
// regarding the given t parameter.
func LerpC(from color.RGBA, to color.RGBA, t float32) color.RGBA {
	return color.RGBA{
		uint8(Lerp(float32(from.R), float32(to.R), t)),
		uint8(Lerp(float32(from.G), float32(to.G), t)),
		uint8(Lerp(float32(from.B), float32(to.B), t)),
		uint8(Lerp(float32(from.A), float32(to.A), t)),
	}
}

// Barycoord computes the barycentric coordinates of a given position
// regards to the given three positions.
func Barycoord(p, t1, t2, t3 Vec2) [3]float32 {
	ap := NewVec3(p.X-t1.X, p.Y-t1.Y, 0)
	ab := NewVec3(t2.X-t1.X, t2.Y-t1.Y, 0)
	ac := NewVec3(t3.X-t1.X, t3.Y-t1.Y, 0)
	bc := NewVec3(t3.X-t2.X, t3.Y-t2.Y, 0)
	bp := NewVec3(p.X-t2.X, p.Y-t2.Y, 0)
	Sabc := ab.Cross(ac).Z
	Sabp := ab.Cross(ap).Z
	Sapc := ap.Cross(ac).Z
	Sbcp := bc.Cross(bp).Z
	w1, w2, w3 := Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc
	return [3]float32{w1, w2, w3}
}

// IsInsideTriangle tests three given vertices and a position p, returns
// true if p is inside the three vertices, or false otherwise.
func IsInsideTriangle(p, v1, v2, v3 Vec2) bool {
	AB := NewVec3(v2.X-v1.X, v2.Y-v1.Y, 0)
	AP := NewVec3(p.X-v1.X, p.Y-v1.Y, 0)
	if AB.Cross(AP).Z < 0 {
		return false
	}
	BC := NewVec3(v3.X-v2.X, v3.Y-v2.Y, 0)
	BP := NewVec3(p.X-v2.X, p.Y-v2.Y, 0)
	if BC.Cross(BP).Z < 0 {
		return false
	}
	CA := NewVec3(v1.X-v3.X, v1.Y-v3.Y, 0)
	CP := NewVec3(p.X-v3.X, p.Y-v3.Y, 0)
	return CA.Cross(CP).Z >= 0
}
