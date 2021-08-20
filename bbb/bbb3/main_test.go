package main_test

import (
	"image"
	"image/color"
	"testing"
)

var col color.RGBA

func loop(img *image.RGBA) {
	for i := 0; i < img.Stride*img.Rect.Dy(); i += 4 {
		s := img.Pix[i : i+4] // here can still be optimized.
		col = color.RGBA{s[0], s[1], s[2], s[3]}
	}
}

func loopByAt(img *image.RGBA) {
	for i := 0; i < img.Bounds().Dy(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			x := img.PixOffset(i, j)
			col = color.RGBA{img.Pix[x], img.Pix[x+1], img.Pix[x+2], img.Pix[x+3]}
		}
	}
}

func BenchmarkLoop(b *testing.B) {
	img := image.NewRGBA(image.Rect(0, 0, 1920, 1080))

	b.Run("loopByAt", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loopByAt(img)
		}
	})
	b.Run("loop", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			loop(img)
		}
	})
}
