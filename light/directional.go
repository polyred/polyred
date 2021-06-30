// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

var (
	_ Source        = &Directional{}
	_ object.Object = &Directional{}
)

// Directional is a directional light source with constant intensity
// at every shading point, and can only cast shadows at a given direction.
type Directional struct {
	math.TransformContext

	pos          math.Vec4
	dir          math.Vec4
	intensity    float64
	color        color.RGBA
	useShadowMap bool
}

type DirectionalOption func(d *Directional)

func WithDirectionalLightPosition(pos math.Vec4) DirectionalOption {
	return func(d *Directional) {
		d.pos = pos
	}
}

func WithDirectionalLightDirection(dir math.Vec4) DirectionalOption {
	return func(d *Directional) {
		d.dir = dir
	}
}

func WithDirectionalLightIntensity(I float64) DirectionalOption {
	return func(d *Directional) {
		d.intensity = I
	}
}

func WithDirectionalLightColor(c color.RGBA) DirectionalOption {
	return func(d *Directional) {
		d.color = c
	}
}

func WithDirectionalLightShadowMap(enable bool) DirectionalOption {
	return func(d *Directional) {
		d.useShadowMap = enable
	}
}

// NewDirectional returns a new directional light
func NewDirectional(opts ...DirectionalOption) Source {
	d := &Directional{
		intensity:    1,
		color:        color.RGBA{255, 255, 255, 255},
		pos:          math.Vec4{},
		dir:          math.NewVec4(0, -1, 0, 0),
		useShadowMap: false,
	}
	for _, opt := range opts {
		opt(d)
	}
	d.dir = d.dir.Unit()
	d.ResetContext()
	return d
}

func (d *Directional) Type() object.Type {
	return object.TypeLight
}

func (d *Directional) Intensity() float64 {
	return d.intensity
}

func (d *Directional) Position() math.Vec4 {
	return d.pos
}

func (d *Directional) Dir() math.Vec4 {
	return d.dir
}

func (d *Directional) Color() color.RGBA {
	return d.color
}

func (d *Directional) CastShadow() bool {
	return d.useShadowMap
}
