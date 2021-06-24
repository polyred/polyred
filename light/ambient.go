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
	_ Environment   = &Ambient{}
	_ object.Object = &Ambient{}
)

type Ambient struct {
	math.TransformContext // not used

	color     color.RGBA
	intensity float64
}

type AmbientOption func(a *Ambient)

func WithAmbientIntensity(I float64) AmbientOption {
	return func(a *Ambient) {
		a.intensity = I
	}
}

func WithAmbientColor(c color.RGBA) AmbientOption {
	return func(a *Ambient) {
		a.color = c
	}
}

func NewAmbient(opts ...AmbientOption) *Ambient {
	a := &Ambient{
		intensity: 0.1,
		color:     color.RGBA{0xff, 0xff, 0xff, 0xff},
	}

	for _, opt := range opts {
		opt(a)
	}
	a.ResetContext()

	return a
}

func (l *Ambient) Type() object.Type {
	return object.TypeLight
}

func (a *Ambient) Color() color.RGBA {
	return a.color
}

func (a *Ambient) Intensity() float64 {
	return a.intensity
}
