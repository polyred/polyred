// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package buffer_test

import (
	"fmt"
	"image"
	"testing"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/texture/imageutil"
)

func mustLoadTexture(path string) *buffer.Texture {
	data := imageutil.MustLoadImage(path)
	return buffer.NewTexture(
		buffer.TextureImage(data),
		buffer.TextureIsoMipmap(true),
	)
}

var (
	data = &image.RGBA{
		Pix: []uint8{
			255, 255, 255, 255, 0, 0, 0, 0,
			0, 0, 0, 0, 255, 255, 255, 255,
		},
		Stride: 8,
		Rect:   image.Rect(0, 0, 2, 2),
	}
	tests = []struct {
		name string
		tex  *buffer.Texture
		u    float32
		v    float32
		lod  float32
		want color.RGBA
	}{
		{"1x1", buffer.NewTexture(), 0, 0, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", buffer.NewTexture(), 1, 1, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", buffer.NewTexture(), 0.5, 0.5, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", buffer.NewTexture(), 0, 0, 0, color.RGBA{255, 255, 255, 255}},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			0, 0, 0, color.RGBA{255, 255, 255, 255},
		},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			0, 1, 0, color.RGBA{0, 0, 0, 0},
		},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			1, 0, 0, color.RGBA{0, 0, 0, 0},
		},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			1, 1, 0, color.RGBA{255, 255, 255, 255},
		},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			1, 1, 1.5, color.RGBA{191, 191, 191, 191},
		},
		{
			"2x2",
			buffer.NewTexture(
				buffer.TextureImage(data),
				buffer.TextureIsoMipmap(true),
			),
			0.5, 0.5, 0, color.RGBA{127, 127, 127, 127},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 0, color.RGBA{99, 142, 9, 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 1, color.RGBA{99, 142, 9, 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 2, color.RGBA{R: 79, G: 119, B: 11, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 3.5, color.RGBA{R: 72, G: 109, B: 13, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 4.5, color.RGBA{R: 75, G: 112, B: 13, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 5.5, color.RGBA{R: 77, G: 114, B: 14, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 6, color.RGBA{R: 78, G: 115, B: 15, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 7, color.RGBA{R: 77, G: 113, B: 15, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 8.5, color.RGBA{R: 79, G: 116, B: 17, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 9.5, color.RGBA{R: 80, G: 116, B: 18, A: 255},
		},
		{
			"1024x1024",
			mustLoadTexture("../internal/testdata/ground.png"),
			0.5, 0.5, 10, color.RGBA{R: 79, G: 116, B: 18, A: 255},
		},
		{
			"pic",
			mustLoadTexture("../internal/testdata/pic.jpg"),
			0.5, 0.5, 0, color.RGBA{R: 253, G: 168, B: 67, A: 255},
		},
	}
)

func TestQuery(t *testing.T) {
	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s-lod-%f", tt.name, tt.lod), func(t *testing.T) {
			got := tt.tex.Query(tt.lod, tt.u, tt.v)
			if !color.Equal(got, tt.want) {
				t.Errorf("#%d want: %+v, got: %+v", i, tt.want, got)
			}
		})
	}
}

func BenchmarkQuery(b *testing.B) {
	for i, tt := range tests {
		ttt := tt
		b.Run(fmt.Sprintf("#%d", i), func(b *testing.B) {
			b.ReportAllocs()
			for j := 0; j < b.N; j++ {
				tt.tex.Query(ttt.u, ttt.v, ttt.lod)
			}
		})
	}
}
