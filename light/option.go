package light

import (
	"image/color"

	"poly.red/math"
)

type Opt func(l interface{})

func Intensity(I float64) Opt {
	return func(l interface{}) {
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
	return func(l interface{}) {
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

func Direction(dir math.Vec3) Opt {
	return func(l interface{}) {
		switch a := l.(type) {
		case *Directional:
			a.dir = dir
		default:
			panic("light: invalid usage of Direction option")
		}
	}
}

func Position(pos math.Vec3) Opt {
	return func(l interface{}) {
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
	return func(l interface{}) {
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