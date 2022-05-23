// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package imageutil_test

import (
	"image/color"
	"image/png"
	"os"
	"testing"

	"poly.red/internal/imageutil"
)

func TestLoadImage(t *testing.T) {
	img := imageutil.MustLoadImage("../../internal/testdata/ground.png", imageutil.GammaCorrect(true))

	f, err := os.Open("../../internal/testdata/golden.png")
	if err != nil {
		t.Fatalf("cannot find golden file")
	}
	want, err := png.Decode(f)
	if err != nil {
		t.Fatalf("cannot decode golden file")
	}

	if !img.Bounds().Eq(want.Bounds()) {
		t.Fatalf("golden image size does not euqal to the loading size")
	}

	for i := 0; i < img.Bounds().Dx(); i++ {
		for j := 0; j < img.Bounds().Dy(); j++ {
			want := want.At(i, j).(color.RGBA)
			got := img.At(i, j).(color.RGBA)
			if want.R != got.R || want.G != got.G || want.B != got.B || want.A != got.A {
				t.Fatalf("pixel (%d, %d) color does not match, want %v, got %v", i, j, want, got)
			}
		}
	}
}

func BenchmarkLoadImage(b *testing.B) {
	b.Run("without-correction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			imageutil.MustLoadImage("../../internal/testdata/ground.png", imageutil.GammaCorrect(false))
		}
	})
	b.Run("with-correction", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			imageutil.MustLoadImage("../../internal/testdata/ground.png", imageutil.GammaCorrect(true))
		}
	})
}
