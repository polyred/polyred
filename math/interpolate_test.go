// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"image/color"
	"testing"

	"changkun.de/x/polyred/math"
)

var (
	global  float64
	globalV math.Vector
	globalC color.RGBA
)

func TestLerp(t *testing.T) {
	if v := math.Lerp(0, 1, 0.5); v != 0.5 {
		t.Fatalf("Lerp want %v, got %v", 0.5, v)
	}
}

func BenchmarkLerp(b *testing.B) {
	t := 0.5
	var r float64
	for i := 0; i < b.N; i++ {
		r = math.Lerp(0, 1, t)
	}
	global = r
}

func TestLerpV(t *testing.T) {
	tt := 0.5
	v1 := math.Vector{0, 0, 0, 1}
	v2 := math.Vector{1, 1, 1, 1}
	want := math.Vector{0.5, 0.5, 0.5, 1}
	if vv := math.LerpV(v1, v2, tt); !vv.Eq(want) {
		t.Fatalf("LerpV want %v, got %v", want, vv)
	}
}

func BenchmarkLerpV(b *testing.B) {
	t := 0.5
	v1 := math.Vector{0, 0, 0, 1}
	v2 := math.Vector{1, 1, 1, 1}
	var r math.Vector
	for i := 0; i < b.N; i++ {
		r = math.LerpV(v1, v2, t)
	}
	globalV = r
}

func TestLerpC(t *testing.T) {
	tt := 0.5
	v1 := color.RGBA{0, 0, 0, 255}
	v2 := color.RGBA{255, 255, 255, 255}
	want := color.RGBA{127, 127, 127, 255}
	if vv := math.LerpC(v1, v2, tt); vv.R != want.R || vv.G != want.G || vv.B != want.B || vv.A != want.A {
		t.Fatalf("LerpC want %v, got %v", want, vv)
	}
}

func BenchmarkLerpC(b *testing.B) {
	t := 0.5
	v1 := color.RGBA{0, 0, 0, 255}
	v2 := color.RGBA{255, 255, 255, 255}
	var r color.RGBA
	for i := 0; i < b.N; i++ {
		r = math.LerpC(v1, v2, t)
	}
	globalC = r
}

var w1, w2, w3 float64

func BenchmarkBarycoord(b *testing.B) {
	v1 := math.NewVector(0, 0, 0, 1)
	v2 := math.NewVector(20, 0, 0, 1)
	v3 := math.NewVector(0, 20, 0, 1)
	b.Run("inside", func(b *testing.B) {
		x := 10
		y := 10
		var ww1, ww2, ww3 float64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ww1, ww2, ww3 = math.Barycoord(math.NewVector(float64(x), float64(y), 0, 1), v1, v2, v3)
		}
		b.StopTimer()
		w1, w2, w3 = ww1, ww2, ww3
	})
	b.Run("outside", func(b *testing.B) {
		x := 20
		y := 20
		var ww1, ww2, ww3 float64
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ww1, ww2, ww3 = math.Barycoord(math.NewVector(float64(x), float64(y), 0, 1), v1, v2, v3)
		}
		b.StopTimer()
		w1, w2, w3 = ww1, ww2, ww3
	})
}
