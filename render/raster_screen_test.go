// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"fmt"
	"image"
	"math/rand"
	"sync/atomic"
	"testing"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/internal/imageutil"
	"poly.red/render"
	"poly.red/shader"
)

func BenchmarkAlphaBlend(b *testing.B) {
	c1 := color.RGBA{128, 128, 128, 128}
	c2 := color.RGBA{128, 128, 128, 128}
	var c color.RGBA
	for i := 0; i < b.N; i++ {
		c = render.AlphaBlend(c1, c2)
	}
	_ = c
}

func TestDrawPixels(t *testing.T) {
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
			render.Size(tt.w, tt.h),
			render.BatchSize(128), // use 128 to make sure all tests covers all cases
		)
		buf := buffer.NewBuffer(image.Rect(0, 0, tt.w, tt.h))

		counter := uint32(0)
		r.DrawFragments(buf, func(frag *primitive.Fragment) color.RGBA {
			atomic.AddUint32(&counter, 1)
			r := uint8(rand.Int())
			g := uint8(rand.Int())
			b := uint8(rand.Int())
			return color.RGBA{R: r, G: g, B: b, A: 255}
		})

		if counter != uint32(tt.w)*uint32(tt.h) {
			t.Errorf("#%d incorrect execution number, want %d, got %d", i, tt.w*tt.h, counter)
			imageutil.Save(buf.Image(), fmt.Sprintf("../internal/examples/out/testdrawpixels-%d.png", i))
		}
	}
}

func BenchmarkDrawFragment(b *testing.B) {
	r := render.NewRenderer(render.Size(1920, 1080))
	buf := buffer.NewBuffer(image.Rect(0, 0, 1920, 1080))
	f := func(f *primitive.Fragment) color.RGBA { return f.Col }
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawFragment(buf, 42, 42, f)
	}
}

func BenchmarkDrawFragment_NonParallel(b *testing.B) {
	r := render.NewRenderer(render.Size(1920, 1080))
	buf := buffer.NewBuffer(image.Rect(0, 0, 1920, 1080))
	f := func(f *primitive.Fragment) color.RGBA { return f.Col }
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for x := 0; x < 1920; x++ {
			for y := 0; y < 1080; y++ {
				r.DrawFragment(buf, x, y, f)
			}
		}
	}
}

func BenchmarkDrawFragment_Parallel(b *testing.B) {
	r := render.NewRenderer(render.Size(1920, 1080))
	buf := buffer.NewBuffer(image.Rect(0, 0, 1920, 1080))
	f := func(f *primitive.Fragment) color.RGBA { return f.Col }
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawFragments(buf, f)
	}
}

func BenchmarkDrawFragments_Size(b *testing.B) {
	w, h := 100, 100
	for i := 1; i < 128; i *= 2 {
		ww, hh := w*i, h*i
		r := render.NewRenderer(render.Size(ww, hh))
		buf := buffer.NewBuffer(image.Rect(0, 0, ww, hh))

		b.Run(fmt.Sprintf("%d-%d", ww, hh), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.DrawFragments(buf, func(frag *primitive.Fragment) color.RGBA {
					return color.RGBA{uint8(frag.X), uint8(frag.X), uint8(frag.Y), uint8(frag.Y)}
				})
			}
		})
	}
}

func BenchmarkDrawFragments_Block_Parallel(b *testing.B) {
	// Notes & Observations:
	//
	// On Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz with 16 cores.
	// If the block size == 32, and the shader computes but simply returns
	// a color to set, a screen pass requires ~2ms. For a 60fps goal,
	// one must optimize the fragment shader down to 14ms.
	ww, hh := 1920, 1080
	for i := 8; i < 1024; i *= 2 {
		buf := buffer.NewBuffer(image.Rect(0, 0, ww, hh))
		r := render.NewRenderer(
			render.Size(ww, hh),
			render.BatchSize(i),
		)
		b.Run(fmt.Sprintf("%d-%d-%d", ww, hh, i), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.DrawFragments(buf, shader.Uniform(color.White))
			}
		})
	}
}

func BenchmarkDrawFragments_Block_NonParallel(b *testing.B) {
	// Notes & Observations:
	//
	// On Intel(R) Core(TM) i9-9900K CPU @ 3.60GHz with 16 cores.
	// The parallelized version only improves ~80% of the computation.
	// See benchmark here:
	//                                    NonParallel    Parallel
	// DrawPixels_Block/1920-1080-8-16    21.6ms ±3%    5.3ms ±5%  -75.41%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-16-16   21.7ms ±3%    4.0ms ±2%  -81.74%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-32-16   21.7ms ±3%    3.8ms ±6%  -82.58%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-64-16   22.2ms ±3%    4.3ms ±1%  -80.62%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-128-16  22.0ms ±5%    4.7ms ±2%  -78.52%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-256-16  22.1ms ±4%    5.9ms ±3%  -73.25%  (p=0.000 n=10+10)
	// DrawPixels_Block/1920-1080-512-16  21.9ms ±3%    8.9ms ±5%  -59.59%  (p=0.000 n=10+10)
	//
	// TODO: optimize the parallel version even better.

	ww, hh := 1920, 1080
	for i := 8; i < 1024; i *= 2 {
		img := image.NewRGBA(image.Rect(0, 0, ww, hh))
		b.Run(fmt.Sprintf("%d-%d-%d", ww, hh, i), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < img.Bounds().Dx(); j++ {
					for k := 0; k < img.Bounds().Dy(); k++ {
						img.SetRGBA(j, k, func() color.RGBA {
							col := color.RGBA{}
							col = img.RGBAAt(j, k)
							col.R = 255
							col.G = 255
							col.B = 255
							col.A = 255
							return col
						}())
					}
				}
			}
		})
	}
}
