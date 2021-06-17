// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"image/color"
	"math/rand"
	"testing"

	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
)

func BenchmarkBlinnPhongShader(b *testing.B) {
	col := color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}
	x := math.Vector{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 1}
	n := math.Vector{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 0}.Unit()
	c := math.Vector{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 1}
	l := []light.Light{
		light.NewPoint(
			light.WithPoingLightItensity(20),
			light.WithPoingLightColor(
				color.RGBA{
					uint8(rand.Int()),
					uint8(rand.Int()),
					uint8(rand.Int()),
					255,
				},
			),
			light.WithPoingLightPosition(
				math.NewVector(rand.Float64(), rand.Float64(), rand.Float64(), 1),
			),
		),
	}

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(material.NewTexture()),
		material.WithBlinnPhongFactors(0.5, 0.6, 200),
		material.WithBlinnPhongShininess(25),
	)

	b.ReportAllocs()
	b.ResetTimer()
	var cc color.RGBA
	for i := 0; i < b.N; i++ {
		cc = mat.FragmentShader(col, x, n, c, l)
	}
	_ = cc
}
