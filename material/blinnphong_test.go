// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"image/color"
	"math/rand"
	"testing"

	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
)

func BenchmarkBlinnPhongShader(b *testing.B) {
	col := color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}
	x := math.Vec4{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 1}
	n := math.Vec4{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 0}.Unit()
	fn := math.Vec4{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 0}.Unit()
	c := math.Vec4{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64(), W: 1}
	l := []light.Source{
		light.NewPoint(
			light.WithPointLightIntensity(20),
			light.WithPointLightColor(
				color.RGBA{
					uint8(rand.Int()),
					uint8(rand.Int()),
					uint8(rand.Int()),
					255,
				},
			),
			light.WithPointLightPosition(
				math.NewVec4(rand.Float64(), rand.Float64(), rand.Float64(), 1),
			),
		),
	}
	a := []light.Environment{
		light.NewAmbient(
			light.WithAmbientIntensity(20),
			light.WithAmbientColor(
				color.RGBA{
					uint8(rand.Int()),
					uint8(rand.Int()),
					uint8(rand.Int()),
					255,
				},
			),
		),
	}

	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(image.NewTexture()),
		material.WithBlinnPhongFactors(0.6, 200),
		material.WithBlinnPhongShininess(25),
	)

	b.ReportAllocs()
	b.ResetTimer()
	var cc color.RGBA
	for i := 0; i < b.N; i++ {
		cc = mat.FragmentShader(col, x, n, fn, c, l, a)
	}
	_ = cc
}
