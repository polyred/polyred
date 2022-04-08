// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ Source                 = &Point{}
	_ object.Object[float32] = &Point{}
)

// Point is a point light
type Point struct {
	math.TransformContext[float32]

	pos          math.Vec3[float32]
	intensity    float32
	color        color.RGBA
	useShadowMap bool
}

// NewPoint returns a new point light
func NewPoint(opts ...Opt) Source {
	l := &Point{
		intensity:    1,
		color:        color.RGBA{255, 255, 255, 255},
		pos:          math.Vec3[float32]{X: 1, Y: 1, Z: 1},
		useShadowMap: false,
	}

	for _, opt := range opts {
		opt(l)
	}
	l.ResetContext()

	return l
}

func (l *Point) Name() string { return "point_light" }

func (l *Point) Type() object.Type {
	return object.TypeLight
}

func (l *Point) Intensity() float32 {
	return l.intensity
}

func (l *Point) Position() math.Vec3[float32] {
	return l.pos
}

func (l *Point) Color() color.RGBA {
	return l.color
}

func (l *Point) CastShadow() bool {
	return l.useShadowMap
}

func (l *Point) AABB() primitive.AABB { return primitive.NewAABB(l.pos) }
