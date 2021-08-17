// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/math"
	"poly.red/object"
)

var (
	_ Source        = &Point{}
	_ object.Object = &Point{}
)

// Point is a point light
type Point struct {
	math.TransformContext

	pos          math.Vec3
	intensity    float64
	color        color.RGBA
	useShadowMap bool
}

// NewPoint returns a new point light
func NewPoint(opts ...Opt) Source {
	l := &Point{
		intensity:    1,
		color:        color.RGBA{255, 255, 255, 255},
		pos:          math.Vec3{X: 1, Y: 1, Z: 1},
		useShadowMap: false,
	}

	for _, opt := range opts {
		opt(l)
	}
	l.ResetContext()

	return l
}

func (l *Point) Type() object.Type {
	return object.TypeLight
}

func (l *Point) Intensity() float64 {
	return l.intensity
}

func (l *Point) Position() math.Vec3 {
	return l.pos
}

func (l *Point) Color() color.RGBA {
	return l.color
}

func (l *Point) CastShadow() bool {
	return l.useShadowMap
}
