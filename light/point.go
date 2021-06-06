// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

type Light interface {
	Itensity() float64
	Position() math.Vector
	Color() color.RGBA
}

// PointLight is a point light
type PointLight struct {
	itensity float64
	color    color.RGBA
	position math.Vector
}

// NewPointLight returns a new point light
func NewPointLight(I float64, c color.RGBA, p math.Vector) Light {
	return &PointLight{
		itensity: I,
		color:    c,
		position: p,
	}
}

func (l *PointLight) Itensity() float64 {
	return l.itensity
}

func (l *PointLight) Position() math.Vector {
	return l.position
}

func (l *PointLight) Color() color.RGBA {
	return l.color
}
