// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

type Light interface {
	Position() math.Vector
}

// PointLight is a point light
type PointLight struct {
	Color    color.RGBA
	position math.Vector
}

// NewPointLight returns a new point light
func NewPointLight(c color.RGBA, p math.Vector) Light {
	return &PointLight{
		Color:    c,
		position: p,
	}
}

func (l *PointLight) Position() math.Vector {
	return l.position
}
