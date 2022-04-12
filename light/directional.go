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
	_ Light                  = &Directional{}
	_ Source                 = &Directional{}
	_ object.Object[float32] = &Directional{}
)

// Directional is a directional light source with constant intensity
// at every shading point, and can only cast shadows at a given direction.
type Directional struct {
	math.TransformContext[float32]

	position     math.Vec3[float32]
	direction    math.Vec3[float32]
	intensity    float32
	color        color.RGBA
	useShadowMap bool
}

// NewDirectional returns a new directional light
func NewDirectional(opts ...Option) Source {
	d := &Directional{
		intensity:    1,
		color:        color.RGBA{255, 255, 255, 255},
		position:     math.Vec3[float32]{},
		direction:    math.NewVec3[float32](0, -1, 0),
		useShadowMap: false,
	}
	for _, opt := range opts {
		opt(d)
	}
	d.direction = d.direction.Unit()
	d.ResetContext()
	return d
}

func (a *Directional) Name() string                 { return "directional_light" }
func (d *Directional) Type() object.Type            { return object.TypeLight }
func (d *Directional) Intensity() float32           { return d.intensity }
func (d *Directional) Position() math.Vec3[float32] { return d.position }
func (d *Directional) Dir() math.Vec3[float32]      { return d.direction }
func (d *Directional) Color() color.RGBA            { return d.color }
func (d *Directional) CastShadow() bool             { return d.useShadowMap }
func (d *Directional) AABB() primitive.AABB         { return primitive.NewAABB(math.NewVec3[float32](0, 0, 0)) }
