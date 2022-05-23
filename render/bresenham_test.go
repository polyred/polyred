// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"image"
	"image/color"
	"testing"

	"poly.red/buffer"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/render"
)

func TestDrawLine(t *testing.T) {
	buf := buffer.NewBuffer(image.Rect(0, 0, 100, 100))
	render.DrawLine(buf, math.NewVec4[float32](0, 0, 0, 1), math.NewVec4[float32](100, 100, 0, 1), color.RGBA{255, 0, 0, 255})
	render.DrawLine(buf, math.NewVec4[float32](0, 100, 0, 1), math.NewVec4[float32](100, 0, 0, 1), color.RGBA{255, 0, 0, 255})

	path := "../internal/examples/out/line.png"
	imageutil.Save(buf.Image(), path)
	t.Logf("render saved at: %v", path)
}
