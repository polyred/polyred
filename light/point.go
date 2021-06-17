// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

type Light interface {
	Itensity() float64
	Position() math.Vector
	Color() color.RGBA
	CastShadow() bool
}

// Point is a point light
type Point struct {
	itensity     float64
	color        color.RGBA
	pos          math.Vector
	useShadowMap bool
}

type PointOption func(l *Point)

func WithPoingLightItensity(I float64) PointOption {
	return func(l *Point) {
		l.itensity = I
	}
}

func WithPoingLightColor(c color.RGBA) PointOption {
	return func(l *Point) {
		l.color = c
	}
}

func WithPoingLightPosition(pos math.Vector) PointOption {
	return func(l *Point) {
		l.pos = pos
	}
}

func WithShadowMap(enable bool) PointOption {
	return func(l *Point) {
		l.useShadowMap = enable
	}
}

// NewPoint returns a new point light
func NewPoint(opts ...PointOption) Light {
	l := &Point{
		itensity:     1,
		color:        color.RGBA{255, 255, 255, 255},
		pos:          math.Vector{},
		useShadowMap: false,
	}

	for _, opt := range opts {
		opt(l)
	}

	return l
}

func (l *Point) Itensity() float64 {
	return l.itensity
}

func (l *Point) Position() math.Vector {
	return l.pos
}

func (l *Point) Color() color.RGBA {
	return l.color
}

func (l *Point) CastShadow() bool {
	return l.useShadowMap
}
