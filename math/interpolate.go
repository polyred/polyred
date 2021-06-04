// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package math

import "image/color"

func Lerp(from float64, to float64, t float64) float64 {
	return from + t*(to-from)
}

func LerpV(from Vector, to Vector, t float64) Vector {
	return Vector{
		Lerp(from.X, to.X, t),
		Lerp(from.Y, to.Y, t),
		Lerp(from.Z, to.Z, t),
		Lerp(from.W, to.W, t),
	}
}

func LerpC(from color.RGBA, to color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		uint8(Lerp(float64(from.R), float64(to.R), t)),
		uint8(Lerp(float64(from.G), float64(to.G), t)),
		uint8(Lerp(float64(from.B), float64(to.B), t)),
		uint8(Lerp(float64(from.A), float64(to.A), t)),
	}
}
