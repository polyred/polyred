package shader

import (
	"poly.red/color"
	"poly.red/geometry/primitive"
)

// Uniform returns a fragment shader that returns
// the given color for every fragments.
func Uniform(c color.RGBA) Fragment {
	return func(_ *primitive.Fragment) color.RGBA {
		return c
	}
}

// Background returns a fragment shader that returns
// the given color for every fragments that its color
// is discarded.
func Background(c color.RGBA) Fragment {
	return func(f *primitive.Fragment) color.RGBA {
		if f.Col == color.Discard {
			return c
		}
		return f.Col
	}
}
