// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"image"
	"image/color"
	"testing"

	"changkun.de/x/ddd/material"
)

func TestBilinear(t *testing.T) {
	data := image.NewRGBA(image.Rectangle{image.Point{0, 0}, image.Point{2, 2}})
	data.Set(0, 0, color.RGBA{255, 255, 255, 255})
	data.Set(0, 1, color.RGBA{0, 0, 0, 0})
	data.Set(1, 0, color.RGBA{0, 0, 0, 0})
	data.Set(1, 1, color.RGBA{255, 255, 255, 255})

	tex := material.NewTexture(data, true)

	col := tex.Query(0, 0, 0)
	if col.R != 255 || col.G != 255 || col.B != 255 || col.A != 255 {
		t.Fatalf("wrong query color")
	}

	col = tex.Query(1, 1, 0)
	if col.R != 255 || col.G != 255 || col.B != 255 || col.A != 255 {
		t.Fatalf("wrong query color, got: %+v", col)
	}
}
