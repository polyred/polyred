// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package math

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
