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
