// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

import (
	"image"
	"image/color"
	"math"
)

// MseDiff computes the mean sequare error difference of two given images.
// If the two given images have different sizes, the function panics.
func MseDiff(img1, img2 image.Image) (*image.RGBA, float64) {
	if !img1.Bounds().Eq(img2.Bounds()) {
		panic("image: incorrect image bounds")
	}

	w := img1.Bounds().Dx()
	h := img1.Bounds().Dy()

	diffImg := image.NewRGBA(img1.Bounds())

	sum := 0.0
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			diff := colorDiff2(img1.At(i, j), img2.At(i, j))
			diffImg.Set(i, j, colorDiff(img1.At(i, j), img2.At(i, j)))
			sum += diff
		}
	}

	return diffImg, sum / float64(w*h)
}

func colorL2diff2(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	dr := (r1 - r2) * (r1 - r2)
	dg := (g1 - g2) * (g1 - g2)
	db := (b1 - b2) * (b1 - b2)

	return float64(dr + dg + db)
}

func colorDiff(c1, c2 color.Color) color.Color {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	dr := uint8(r1 - r2)
	dg := uint8(g1 - g2)
	db := uint8(b1 - b2)
	return color.RGBA{R: dr, G: dg, B: db, A: 255}
}

// colorDiff2 implements https://www.compuphase.com/cmetric.htm
func colorDiff2(c1, c2 color.Color) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	dr := (r1 - r2) * (r1 - r2)
	dg := (g1 - g2) * (g1 - g2)
	db := (b1 - b2) * (b1 - b2)
	t := (r1 + r2) / 2

	v := 2*dr + 4*dg + 3*db + t*(dr-db)/256
	return math.Sqrt(float64(v))
}
