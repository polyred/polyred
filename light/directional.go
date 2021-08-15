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
	_ Source        = &Directional{}
	_ object.Object = &Directional{}
)

// Directional is a directional light source with constant intensity
// at every shading point, and can only cast shadows at a given direction.
type Directional struct {
	math.TransformContext

	pos          math.Vec3
	dir          math.Vec3
	intensity    float64
	color        color.RGBA
	useShadowMap bool
}

// NewDirectional returns a new directional light
func NewDirectional(opts ...Opt) Source {
	d := &Directional{
		intensity:    1,
		color:        color.RGBA{255, 255, 255, 255},
		pos:          math.Vec3{},
		dir:          math.NewVec3(0, -1, 0),
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

func (d *Directional) Position() math.Vec3 {
	return d.pos
}

func (d *Directional) Dir() math.Vec3 {
	return d.dir
}

func (d *Directional) Color() color.RGBA {
	return d.color
}

func (d *Directional) CastShadow() bool {
	return d.useShadowMap
}
