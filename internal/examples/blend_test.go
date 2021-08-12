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
	"poly.red/texture"

	"poly.red/internal/utils"
)

func TestBlending(t *testing.T) {
	img1 := texture.MustLoadImage("../testdata/src1.png")
	img2 := texture.MustLoadImage("../testdata/src2.png")
	want := texture.MustLoadImage("../testdata/blend.png")
	r := render.NewRenderer(
		render.Size(img1.Bounds().Dx(), img1.Bounds().Dy()),
		render.Blending(render.AlphaBlend),
	)
	dst := img1
	r.ScreenPass(dst, func(f primitive.Fragment) color.RGBA {
		return img2.RGBAAt(f.X, f.Y)
	})

	utils.Save(dst, "./out/dst.png")
	diff, num := texture.MseDiff(want, dst)
	utils.Save(diff, "./out/diff.png")
	fmt.Println("total diff: ", num, img1.Bounds().Dx()*img1.Bounds().Dy())
}
