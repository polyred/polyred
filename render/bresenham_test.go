// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"image"
	"image/color"
	"testing"

	"poly.red/math"
	"poly.red/render"
	"poly.red/texture"
	"poly.red/texture/imageutil"
)

func TestDrawLine(t *testing.T) {
	buf := texture.NewBuffer(image.Rect(0, 0, 100, 100))

	render.DrawLine(buf, math.NewVec4(0, 0, 0, 1), math.NewVec4(100, 100, 0, 1), color.RGBA{255, 0, 0, 255})
	render.DrawLine(buf, math.NewVec4(0, 100, 0, 1), math.NewVec4(100, 0, 0, 1), color.RGBA{255, 0, 0, 255})

	imageutil.Save(buf.Image(), "line.png")
}
