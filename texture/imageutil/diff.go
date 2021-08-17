// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package imageutil

import (
	"image"
	"image/color"
	"math"
)

// DiffKernel is a kernel function that consumes two colors and returns
// their difference.
type DiffKernel func(c1, c2 color.RGBA) float64

// Diff computes the error difference of two given images with respect to the
// given kernel function.
//
// If the two given images have different sizes, the function panics.
func Diff(img1, img2 *image.RGBA, kernel DiffKernel) (*image.RGBA, float64) {
	if !img1.Bounds().Eq(img2.Bounds()) {
		panic("image: incorrect image bounds")
	}

	w := img1.Bounds().Dx()
	h := img1.Bounds().Dy()

	diffImg := image.NewRGBA(img1.Bounds())

	sum := 0.0
	for i := 0; i < w; i++ {
		for j := 0; j < h; j++ {
			score := kernel(img1.RGBAAt(i, j), img2.RGBAAt(i, j)) // colorDiff2
			diffImg.SetRGBA(i, j, colorDiff(img1.RGBAAt(i, j), img2.RGBAAt(i, j)))
			sum += score
		}
	}

	return diffImg, sum / float64(w*h)
}

func MseKernel(c1, c2 color.RGBA) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()

	dr := (r1 - r2) * (r1 - r2)
	dg := (g1 - g2) * (g1 - g2)
	db := (b1 - b2) * (b1 - b2)

	return float64(dr + dg + db)
}

// TODO: figure out how to export diff image?
// Maybe a kernel: func(color.RGBA, color.RGBA) (color.RGBA, float64)?

func colorDiff(c1, c2 color.RGBA) color.RGBA {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	dr := uint8(r1 - r2)
	dg := uint8(g1 - g2)
	db := uint8(b1 - b2)
	return color.RGBA{R: dr, G: dg, B: db, A: 255}
}

// colorDiff2 implements https://www.compuphase.com/cmetric.htm
func colorDiff2(c1, c2 color.RGBA) float64 {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	dr := (r1 - r2) * (r1 - r2)
	dg := (g1 - g2) * (g1 - g2)
	db := (b1 - b2) * (b1 - b2)
	t := (r1 + r2) / 2

	v := 2*dr + 4*dg + 3*db + t*(dr-db)/256
	return math.Sqrt(float64(v))
}
