// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"fmt"
	"image"
	"testing"

	"changkun.de/x/ddd/color"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/material"
)

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
		tex  *material.Texture
		u    float64
		v    float64
		lod  float64
		want color.RGBA
	}{
		{"1x1", material.NewTexture(), 0, 0, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", material.NewTexture(), 1, 1, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", material.NewTexture(), 0.5, 0.5, 0, color.RGBA{255, 255, 255, 255}},
		{"1x1", material.NewTexture(), 0, 0, 0, color.RGBA{255, 255, 255, 255}},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			0, 0, 0, color.RGBA{255, 255, 255, 255},
		},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			0, 1, 0, color.RGBA{0, 0, 0, 0},
		},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			1, 0, 0, color.RGBA{0, 0, 0, 0},
		},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			1, 1, 0, color.RGBA{255, 255, 255, 255},
		},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			1, 1, 1.5, color.RGBA{191, 191, 191, 191},
		},
		{
			"2x2",
			material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
			),
			0.5, 0.5, 0, color.RGBA{127, 127, 127, 127},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 0, color.RGBA{99, 142, 9, 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 1, color.RGBA{99, 142, 9, 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 2, color.RGBA{R: 67, G: 107, B: 11, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 3.5, color.RGBA{R: 76, G: 109, B: 19, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 4.5, color.RGBA{R: 92, G: 134, B: 22, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 5.5, color.RGBA{R: 77, G: 117, B: 17, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 6, color.RGBA{R: 67, G: 107, B: 11, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 7, color.RGBA{R: 67, G: 107, B: 11, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 8.5, color.RGBA{R: 65, G: 106, B: 12, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 9.5, color.RGBA{R: 61, G: 102, B: 11, A: 255},
		},
		{
			"1024x1024",
			io.MustLoadTexture("../testdata/ground.png"),
			0.5, 0.5, 10, color.RGBA{R: 66, G: 105, B: 12, A: 255},
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
			for j := 0; j < b.N; j++ {
				tt.tex.Query(ttt.u, ttt.v, ttt.lod)
			}
		})
	}
}
