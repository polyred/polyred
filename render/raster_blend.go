// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import "image/color"

// BlendFunc is a blending function for two given colors and returns
// the resulting color.
type BlendFunc func(dst, src color.RGBA) color.RGBA

// AlphaBlend performs alpha blending for pre-multiplied alpha RGBA colors
func AlphaBlend(dst, src color.RGBA) color.RGBA {
	// FIXME: there is an overflow
	sr, sg, sb, sa := uint32(src.R), uint32(src.G), uint32(src.B), uint32(src.A)
	dr, dg, db, da := uint32(dst.R), uint32(dst.G), uint32(dst.B), uint32(dst.A)

	// dr, dg, db and da are all 8-bit color at the moment, ranging in [0,255].
	// We work in 16-bit color, and so would normally do:
	// dr |= dr << 8
	// and similarly for dg, db and da, but instead we multiply a
	// (which is a 16-bit color, ranging in [0,65535]) by 0x101.
	// This yields the same result, but is fewer arithmetic operations.
	a := (0xffff - sa) * 0x101

	r := sr + dr*a/0xffff
	g := sg + dg*a/0xffff
	b := sb + db*a/0xffff
	aa := sa + da*a/0xffff
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(aa)}
}
