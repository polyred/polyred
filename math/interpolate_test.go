// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"image/color"
	"testing"

	"changkun.de/x/ddd/math"
)

func TestLerp(t *testing.T) {
	if v := math.Lerp(0, 1, 0.5); v != 0.5 {
		t.Fatalf("Lerp want %v, got %v", 0.5, v)
	}
}

func BenchmarkLerp(b *testing.B) {
	t := 0.5
	for i := 0; i < b.N; i++ {
		math.Lerp(0, 1, t)
	}
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
	for i := 0; i < b.N; i++ {
		math.LerpV(v1, v2, t)
	}
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
	for i := 0; i < b.N; i++ {
		math.LerpC(v1, v2, t)
	}
}
