// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/math"
)

type Opt func(l any)

func Intensity(I float32) Opt {
	return func(l any) {
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

func Color(c color.RGBA) Opt {
	return func(l any) {
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

func Direction(dir math.Vec3[float32]) Opt {
	return func(l any) {
		switch a := l.(type) {
		case *Directional:
			a.dir = dir
		default:
			panic("light: invalid usage of Direction option")
		}
	}
}

func Position(pos math.Vec3[float32]) Opt {
	return func(l any) {
		switch a := l.(type) {
		case *Directional:
			a.pos = pos
		case *Point:
			a.pos = pos
		default:
			panic("light: invalid usage of Position option")
		}
	}
}

func CastShadow(enable bool) Opt {
	return func(l any) {
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
