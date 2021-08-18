// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package example_test

import (
	"fmt"
	"image/color"
	"testing"

	"poly.red/geometry/primitive"
	"poly.red/render"
	"poly.red/texture/imageutil"
)

func TestBlending(t *testing.T) {
	img1 := imageutil.MustLoadImage("../testdata/src1.png")
	img2 := imageutil.MustLoadImage("../testdata/src2.png")
	want := imageutil.MustLoadImage("../testdata/blend.png")
	r := render.NewRenderer(
		render.Size(img1.Bounds().Dx(), img1.Bounds().Dy()),
		render.Blending(render.AlphaBlend),
	)
	dst := img1
	r.DrawPixels(dst, func(f primitive.Fragment) color.RGBA {
		return img2.RGBAAt(f.X, f.Y)
	})

	imageutil.Save(dst, "./out/dst.png")
	diff, num := imageutil.Diff(want, dst, imageutil.MseKernel)
	imageutil.Save(diff, "./out/diff.png")
	fmt.Println("total diff: ", num, img1.Bounds().Dx()*img1.Bounds().Dy())
}
