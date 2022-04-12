// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/math"
)

type Option func(l Light)

func Intensity(I float32) Option {
	return func(l Light) {
		switch a := l.(type) {
		case *Ambient:
			a.intensity = I
		case *Directional:
			a.intensity = I
		case *Point:
			a.intensity = I
		default:
			panic("light: invalid usage of Intensity option")
		}
	}
}

func Color(c color.RGBA) Option {
	return func(l Light) {
		switch a := l.(type) {
		case *Ambient:
			a.color = c
		case *Directional:
			a.color = c
		case *Point:
			a.color = c
		default:
			panic("light: invalid usage of Color option")
		}
	}
}

func Direction(dir math.Vec3[float32]) Option {
	return func(l Light) {
		switch a := any(l).(type) {
		case *Directional:
			a.direction = dir
		default:
			panic("light: invalid usage of Direction option")
		}
	}
}

func Position(pos math.Vec3[float32]) Option {
	return func(l Light) {
		switch a := l.(type) {
		case *Directional:
			a.position = pos
		case *Point:
			a.position = pos
		default:
			panic("light: invalid usage of Position option")
		}
	}
}

func CastShadow(enable bool) Option {
	return func(l Light) {
		switch a := l.(type) {
		case *Directional:
			a.useShadowMap = enable
		case *Point:
			a.useShadowMap = enable
		default:
			panic("light: invalid usage of CastShadow option")
		}
	}
}
