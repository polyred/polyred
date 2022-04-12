// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ Light                  = &Ambient{}
	_ Environment            = &Ambient{}
	_ object.Object[float32] = &Ambient{}
)

type Ambient struct {
	math.TransformContext[float32] // not used

	color     color.RGBA
	intensity float32
}

func NewAmbient(opts ...Option) Environment {
	a := &Ambient{
		intensity: 0.1,
		color:     color.White,
	}

	for _, opt := range opts {
		opt(a)
	}
	a.ResetContext()
	return a
}

func (a *Ambient) Name() string         { return "ambient_light" }
func (a *Ambient) Type() object.Type    { return object.TypeLight }
func (a *Ambient) Color() color.RGBA    { return a.color }
func (a *Ambient) Intensity() float32   { return a.intensity }
func (a *Ambient) AABB() primitive.AABB { return primitive.NewAABB(math.NewVec3[float32](0, 0, 0)) }
