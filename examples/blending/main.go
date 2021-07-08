package main

import (
	"fmt"
	"image/color"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/utils"
)

func main() {
	img1 := io.MustLoadImage("../../testdata/src1.png")
	img2 := io.MustLoadImage("../../testdata/src2.png")
	want := io.MustLoadImage("./dst.png")
	r := render.NewRenderer(
		render.WithSize(img1.Bounds().Dx(), img1.Bounds().Dy()),
		render.WithBlendFunc(render.AlphaBlend),
	)
	dst := img1
	r.ScreenPass(dst, func(f primitive.Fragment) color.RGBA {
		return img2.RGBAAt(f.X, f.Y)
	})

	utils.Save(dst, "dst.png")
	diff, num := image.MseDiff(want, dst)
	utils.Save(diff, "diff.png")
	fmt.Println("total diff: ", num, img1.Bounds().Dx()*img1.Bounds().Dy())
}
