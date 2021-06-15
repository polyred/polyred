// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package color

import (
	"fmt"
	"image/color"
	"strings"
)

type RGBA = color.RGBA

func FromHex(x string) color.RGBA {
	x = strings.Trim(x, "#")
	var r, g, b, a uint8
	a = 255
	switch len(x) {
	case 3:
		fmt.Sscanf(x, "%1x%1x%1x", &r, &g, &b)
		r = (r << 4) | r
		g = (g << 4) | g
		b = (b << 4) | b
	case 4:
		fmt.Sscanf(x, "%1x%1x%1x%1x", &r, &g, &b, &a)
		r = (r << 4) | r
		g = (g << 4) | g
		b = (b << 4) | b
		a = (a << 4) | a
	case 6:
		fmt.Sscanf(x, "%02x%02x%02x", &r, &g, &b)
	case 8:
		fmt.Sscanf(x, "%02x%02x%02x%02x", &r, &g, &b, &a)
	}
	return color.RGBA{r, g, b, 0xff}
}

func Equal(c1, c2 color.RGBA) bool {
	if c1.R == c2.R && c1.G == c2.G && c1.B == c2.B && c1.A == c2.A {
		return true
	}
	return false
}
