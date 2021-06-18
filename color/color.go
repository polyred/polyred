// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package color

import (
	"fmt"
	"image/color"
	"math"
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

// Linear2Gamma applies gamma correction to v and lies in [0.0, 1.0].
func Linear2Gamma(v float64) float64 {
	return math.Min(math.Max(ConvertLinear2sRGB(v), 0), 1)
}

// Gamma2Linear applies inverse gamma correction v and lies in [0.0, 1.0].
func Gamma2Linear(v float64) float64 {
	return math.Min(math.Max(ConvertSRGB2Linear(v), 0), 1)
}

// ConvertLinear2sRGB is a sRGB encoder
func ConvertLinear2sRGB(v float64) float64 {
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1/2.4) - 0.055
	}
	return v
}

// ConvertSRGB2Linear is a sRGB decoder
func ConvertSRGB2Linear(v float64) float64 {
	if v <= 0.04045 {
		v /= 12.92
	} else {
		v = math.Pow((v+0.055)/1.055, 2.4)
	}
	return v
}
