// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material_test

import (
	"math/rand"
	"testing"

	"poly.red/buffer"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
)

func BenchmarkBlinnPhongShader(b *testing.B) {
	x := math.NewRandVec3[float32]().ToVec4(1)
	n := math.NewRandVec3[float32]().ToVec4(0).Unit()
	fn := math.NewRandVec3[float32]().ToVec4(0).Unit()
	c := math.NewRandVec3[float32]()
	l := []light.Source{
		light.NewPoint(
			light.Intensity(20),
			light.Color(color.RGBA{
				uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255}),
			light.Position(math.NewRandVec3[float32]()),
		),
	}
	a := []light.Environment{
		light.NewAmbient(
			light.Intensity(20),
			light.Color(color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255}),
		),
	}

	m := material.NewBlinnPhong(
		material.Texture(buffer.NewTexture()),
		material.Kdiff(color.FromValue(0.6, 0.6, 0.6, 1.0)),
		material.Kspec(color.FromValue(0.6, 0.6, 0.6, 1.0)),
		material.Shininess(25),
	)

	frag := primitive.Fragment{
		U:        1,
		V:        1,
		Du:       1,
		Dv:       1,
		Nor:      n,
		AttrFlat: map[primitive.AttrName]any{},
	}
	frag.AttrFlat["fN"] = fn
	frag.AttrFlat["Pos"] = x
	info := buffer.Fragment{Ok: true, Fragment: frag}
	b.ReportAllocs()
	b.ResetTimer()
	var cc color.RGBA
	for i := 0; i < b.N; i++ {
		cc = m.FragmentShader(info, c, l, a)
	}
	_ = cc
}
