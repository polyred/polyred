// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

// PointLight is a point light
type PointLight struct {
	Color    color.RGBA
	Position math.Vector
}

// NewPointLight returns a new point light
func NewPointLight(c color.RGBA, p math.Vector) PointLight {
	return PointLight{
		Color:    c,
		Position: p,
	}
}
