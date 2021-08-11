package main

import (
	"fmt"
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/render"
	"poly.red/texture"

	"poly.red/internal/utils"
)

func main() {
	img1 := texture.MustLoadImage("../../testdata/src1.png")
	img2 := texture.MustLoadImage("../../testdata/src2.png")
	want := texture.MustLoadImage("./dst.png")
	r := render.NewRenderer(
		render.WithSize(img1.Bounds().Dx(), img1.Bounds().Dy()),
		render.WithBlendFunc(render.AlphaBlend),
	)
	dst := img1
	r.ScreenPass(dst, func(f primitive.Fragment) color.RGBA {
		return img2.RGBAAt(f.X, f.Y)
	})

	utils.Save(dst, "dst.png")
	diff, num := texture.MseDiff(want, dst)
	utils.Save(diff, "diff.png")
	fmt.Println("total diff: ", num, img1.Bounds().Dx()*img1.Bounds().Dy())
}
