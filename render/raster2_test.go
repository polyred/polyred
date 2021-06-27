// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"changkun.de/x/polyred/render"
)

func TestScreenPass(t *testing.T) {
	// smaller than concurrent size
	w, h := 100, 100
	r := render.NewRenderer(
		render.WithSize(w, h),
		render.WithConcurrency(128),
	)
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	r.ScreenPass(img, func(x, y int) color.RGBA {
		return color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x), A: uint8(y)}
	})
	// utils.Save(img, "1.png")

	// w is smaller than concurrent size
	w, h = 100, 200
	r = render.NewRenderer(
		render.WithSize(w, h),
		render.WithConcurrency(128),
	)
	img = image.NewRGBA(image.Rect(0, 0, w, h))

	r.ScreenPass(img, func(x, y int) color.RGBA {
		return color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x), A: uint8(y)}
	})
	// utils.Save(img, "2.png")

	// h is smaller than concurrent size
	w, h = 200, 100
	r = render.NewRenderer(
		render.WithSize(w, h),
		render.WithConcurrency(128),
	)
	img = image.NewRGBA(image.Rect(0, 0, w, h))

	r.ScreenPass(img, func(x, y int) color.RGBA {
		return color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x), A: uint8(y)}
	})
	// utils.Save(img, "3.png")

	// both greater than concurrent size
	w, h = 200, 200
	r = render.NewRenderer(
		render.WithSize(w, h),
		render.WithConcurrency(128),
	)
	img = image.NewRGBA(image.Rect(0, 0, w, h))

	r.ScreenPass(img, func(x, y int) color.RGBA {
		return color.RGBA{R: uint8(x), G: uint8(y), B: uint8(x), A: uint8(y)}
	})
	// utils.Save(img, "4.png")
}

func BenchmarkScreenPass_Size(b *testing.B) {
	w, h := 100, 100
	for i := 1; i < 128; i *= 2 {
		ww, hh := w*i, h*i
		r := render.NewRenderer(
			render.WithSize(ww, hh),
			render.WithConcurrency(128),
		)
		img := image.NewRGBA(image.Rect(0, 0, ww, hh))

		b.Run(fmt.Sprintf("%d-%d", ww, hh), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.ScreenPass(img, func(x, y int) color.RGBA {
					return color.RGBA{uint8(x), uint8(y), uint8(x), uint8(y)}
				})
			}
		})
	}
}

func BenchmarkScreenPass_Block(b *testing.B) {
	ww, hh := 1920, 1080
	for i := 1; i < 1024; i *= 2 {
		img := image.NewRGBA(image.Rect(0, 0, ww, hh))
		r := render.NewRenderer(
			render.WithSize(ww, hh),
			render.WithConcurrency(int32(i)),
		)
		b.Run(fmt.Sprintf("%d-%d-%d", ww, hh, i), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.ScreenPass(img, func(x, y int) color.RGBA {
					return color.RGBA{uint8(x), uint8(y), uint8(x), uint8(y)}
				})
			}
		})
	}
}
