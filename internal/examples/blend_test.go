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
	"poly.red/texture/buffer"
	"poly.red/texture/imageutil"
)

func TestBlending(t *testing.T) {
	img1 := imageutil.MustLoadImage("../testdata/src1.png")
	img2 := imageutil.MustLoadImage("../testdata/src2.png")
	want := imageutil.MustLoadImage("../testdata/blend.png")

	buf1 := buffer.NewBuffer(img1.Rect)
	for i := 0; i < buf1.Bounds().Dx(); i++ {
		for j := 0; j < buf1.Bounds().Dy(); j++ {
			buf1.Set(i, j, buffer.Fragment{Ok: true, Fragment: primitive.Fragment{
				X: i, Y: j, Col: img1.RGBAAt(i, img1.Bounds().Dy()-j-1),
			}})
		}
	}

	buf2 := buffer.NewBuffer(img2.Rect)
	for i := 0; i < buf2.Bounds().Dx(); i++ {
		for j := 0; j < buf2.Bounds().Dy(); j++ {
			buf2.Set(i, j, buffer.Fragment{Ok: true, Fragment: primitive.Fragment{
				X: i, Y: j, Col: img2.RGBAAt(i, img2.Bounds().Dy()-j-1),
			}})
		}
	}

	r := render.NewRenderer(
		render.Size(img1.Bounds().Dx(), img1.Bounds().Dy()),
		render.Blending(render.AlphaBlend),
	)

	r.DrawFragments(buf1, func(f primitive.Fragment) color.RGBA {
		return buf2.At(f.X, f.Y).Col
	})

	imageutil.Save(buf1.Image(), "./out/dst.png")
	diff, num := imageutil.Diff(want, buf1.Image(), imageutil.MseKernel)
	imageutil.Save(diff, "./out/diff.png")
	fmt.Println("total diff: ", num, img1.Bounds().Dx()*img1.Bounds().Dy())
}
