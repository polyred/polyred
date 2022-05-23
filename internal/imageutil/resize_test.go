// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package imageutil_test

import (
	"image"
	"testing"

	"poly.red/internal/imageutil"
)

var dst *image.RGBA

func BenchmarkResize(b *testing.B) {
	b.Run("ScaleDown2x", func(b *testing.B) {
		img := imageutil.MustLoadImage("../../internal/examples/out/shadow.png")
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dst = imageutil.Resize(img.Bounds().Dx()/2, img.Bounds().Dy()/2, img)
		}
	})
}
