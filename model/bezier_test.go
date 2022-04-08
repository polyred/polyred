// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model_test

import (
	"image"
	"testing"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
)

func TestBazier(t *testing.T) {
	b := model.NewBezierCurve(&primitive.Vertex{
		Pos: math.NewVec4[float32](0, 0, 0, 1),
	}, &primitive.Vertex{
		Pos: math.NewVec4[float32](1, 0, 0, 1),
	}, &primitive.Vertex{
		Pos: math.NewVec4[float32](0, 1, 0, 1),
	}, &primitive.Vertex{
		Pos: math.NewVec4[float32](1, 1, 0, 1),
	})

	scale := float32(1000)
	buf := buffer.NewBuffer(image.Rect(0, 0, int(scale), int(scale)))
	prev := b.At(float32(0)).Scale(scale, scale, scale, 1)
	for i := 0.01; i <= 1; i += 0.001 {
		next := b.At(float32(i)).Scale(scale, scale, scale, 1)
		render.DrawLine(buf, prev, next, color.Red)
		prev = next
	}

	imageutil.Save(buf.Image(), "../internal/examples/out/bazier.png")
}
