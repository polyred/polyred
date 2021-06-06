// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package utils

import (
	"image/color"
)

func ScaleDown2x(width, height float64, buf []color.RGBA) []color.RGBA {
	w := int(0.5 * width)
	h := int(0.5 * height)
	ret := make([]color.RGBA, w*h)
	for i := 0; i < w*2; i += 2 {
		for j := 0; j < h*2; j += 2 {
			c1 := buf[i+j*w*2]
			c2 := buf[i+1+j*w*2]
			c3 := buf[i+(j+1)*w*2]
			c4 := buf[i+1+(j+1)*w*2]
			r := (float64(c1.R) + float64(c2.R) + float64(c3.R) + float64(c4.R)) / 4
			g := (float64(c1.G) + float64(c2.G) + float64(c3.G) + float64(c4.G)) / 4
			b := (float64(c1.B) + float64(c2.B) + float64(c3.B) + float64(c4.B)) / 4
			ret[(i+j*w)/2] = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
		}
	}
	return ret
}
