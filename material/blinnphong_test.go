// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"image/color"
	"math/rand"
	"testing"

	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/texture"
)

func BenchmarkBlinnPhongShader(b *testing.B) {
	col := color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}
	x := math.NewRandVec3().ToVec4(1)
	n := math.NewRandVec3().ToVec4(0).Unit()
	fn := math.NewRandVec3().ToVec4(0).Unit()
	c := math.NewRandVec3().ToVec4(1)
	l := []light.Source{
		light.NewPoint(
			light.Intensity(20),
			light.Color(color.RGBA{
				uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255}),
			light.Position(math.NewRandVec3()),
		),
	}
	a := []light.Environment{
		light.NewAmbient(
			light.Intensity(20),
			light.Color(color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255}),
		),
	}

	mat := material.NewBlinnPhong(
		material.Texture(texture.NewTexture()),
		material.Kdiff(0.6), material.Kspec(200),
		material.Shininess(25),
	)

	b.ReportAllocs()
	b.ResetTimer()
	var cc color.RGBA
	for i := 0; i < b.N; i++ {
		cc = mat.FragmentShader(col, x, n, fn, c, l, a)
	}
	_ = cc
}
