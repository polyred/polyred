// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"fmt"
	"image"
	"math/rand"
	"sync/atomic"
	"testing"

	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/utils"
)

func TestScreenPass(t *testing.T) {
	tests := []struct {
		w int
		h int
	}{
		// smaller than concurrent size
		{100, 100},
		// w is smaller than concurrent size
		{100, 200},
		// h is smaller than concurrent size
		{200, 100},
		// both greater than concurrent size
		{200, 200},
	}

	for i, tt := range tests {
		r := render.NewRenderer(
			render.WithSize(tt.w, tt.h),
			render.WithConcurrency(128),
		)
		img := image.NewRGBA(image.Rect(0, 0, tt.w, tt.h))

		counter := uint32(0)
		r.ScreenPass(img, func(x, y int, col color.RGBA) color.RGBA {
			atomic.AddUint32(&counter, 1)
			r := uint8(rand.Int())
			g := uint8(rand.Int())
			b := uint8(rand.Int())
			return color.RGBA{R: r, G: g, B: b, A: 255}
		})

		if counter != uint32(tt.w)*uint32(tt.h) {
			t.Errorf("#%d incorrect execution number, want %d, got %d", i, tt.w*tt.h, counter)
			utils.Save(img, fmt.Sprintf("%d.png", i))
		}
	}
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
				r.ScreenPass(img, func(x, y int, col color.RGBA) color.RGBA {
					return color.RGBA{uint8(x), uint8(y), uint8(x), uint8(y)}
				})
			}
		})
	}
}

func BenchmarkScreenPass_Block(b *testing.B) {
	// Notes & Observations:
	//
	// On Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz with 16 cores.
	// If the block size == 32, and the shader computes but simply returns
	// a color to set, a screen pass requires ~2ms. For a 60fps goal,
	// one must optimize the fragment shader down to 14ms.
	ww, hh := 1920, 1080
	for i := 1; i <= 1024; i *= 2 {
		img := image.NewRGBA(image.Rect(0, 0, ww, hh))
		r := render.NewRenderer(
			render.WithSize(ww, hh),
			render.WithConcurrency(int32(i)),
		)
		b.Run(fmt.Sprintf("%d-%d-%d", ww, hh, i), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.ScreenPass(img, func(x, y int, col color.RGBA) color.RGBA {
					return color.RGBA{255, 255, 255, 255}
				})
			}
		})
	}
}
