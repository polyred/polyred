// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader

import (
	"poly.red/color"
	"poly.red/geometry/primitive"
)

// GammaCorrection is a fragment shader that processes a given fragment.
func GammaCorrection(frag *primitive.Fragment) color.RGBA {
	r := uint8(color.FromLinear2sRGB(float32(frag.Col.R)/0xff)*0xff + 0.5)
	g := uint8(color.FromLinear2sRGB(float32(frag.Col.G)/0xff)*0xff + 0.5)
	b := uint8(color.FromLinear2sRGB(float32(frag.Col.B)/0xff)*0xff + 0.5)
	return color.RGBA{r, g, b, frag.Col.A}
}
