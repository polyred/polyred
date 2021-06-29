// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import (
	"image/color"
)

// Lerp computes a linear interpolation between two given numbers
// regarding the given t parameter.
func Lerp(from float64, to float64, t float64) float64 {
	return from + t*(to-from)
}

// LerpV computes a linear interpolation between two given vectors
// regarding the given t parameter.
func LerpV(from Vector, to Vector, t float64) Vector {
	return Vector{
		Lerp(from.X, to.X, t),
		Lerp(from.Y, to.Y, t),
		Lerp(from.Z, to.Z, t),
		Lerp(from.W, to.W, t),
	}
}

// LerpC computes a linear interpolation between two given colors
// regarding the given t parameter.
func LerpC(from color.RGBA, to color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		uint8(Lerp(float64(from.R), float64(to.R), t)),
		uint8(Lerp(float64(from.G), float64(to.G), t)),
		uint8(Lerp(float64(from.B), float64(to.B), t)),
		uint8(Lerp(float64(from.A), float64(to.A), t)),
	}
}

// Barycoord computes the barycentric coordinates of a given position
// regards to the given three positions.
func Barycoord(x, y float64, t1, t2, t3 Vector) (w1, w2, w3 float64) {
	ap := NewVector(x-t1.X, y-t1.Y, 0, 0)
	ab := NewVector(t2.X-t1.X, t2.Y-t1.Y, 0, 0)
	ac := NewVector(t3.X-t1.X, t3.Y-t1.Y, 0, 0)
	bc := NewVector(t3.X-t2.X, t3.Y-t2.Y, 0, 0)
	bp := NewVector(x-t2.X, y-t2.Y, 0, 0)
	Sabc := ab.Cross(ac).Z
	Sabp := ab.Cross(ap).Z
	Sapc := ap.Cross(ac).Z
	Sbcp := bc.Cross(bp).Z
	w1, w2, w3 = Sbcp/Sabc, Sapc/Sabc, Sabp/Sabc
	return
}

// IsInsideTriangle tests three given vertices and a position p, returns
// true if p is inside the three vertices, or false otherwise.
func IsInsideTriangle(v1, v2, v3, p Vector) bool {
	AB := NewVector(v2.X, v2.Y, 0, 1).Sub(NewVector(v1.X, v1.Y, 0, 1))
	AP := p.Sub(NewVector(v1.X, v1.Y, 0, 1))
	if AB.Cross(AP).Z < 0 {
		return false
	}
	BC := NewVector(v3.X, v3.Y, 0, 1).Sub(NewVector(v2.X, v2.Y, 0, 1))
	BP := p.Sub(NewVector(v2.X, v2.Y, 0, 1))
	if BC.Cross(BP).Z < 0 {
		return false
	}
	CA := NewVector(v1.X, v1.Y, 0, 1).Sub(NewVector(v3.X, v3.Y, 0, 1))
	CP := p.Sub(NewVector(v3.X, v3.Y, 0, 1))
	return CA.Cross(CP).Dot(NewVector(0, 0, 1, 0)) >= 0
}
