// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package color_test

import (
	"math/rand"
	"testing"

	"poly.red/color"
	"poly.red/math"
)

func TestColor(t *testing.T) {
	tests := []struct {
		col string
	}{
		{col: "#ffffffff"},
		{col: "#ffffff"},
		{col: "#ffff"},
		{col: "#fff"},
	}

	for _, tt := range tests {
		col := color.FromHex(tt.col)
		if !color.Equal(col, color.White) {
			t.Fatalf("unexpected color from hex, got %v, want %v", col, color.White)
		}

		if color.Equal(col, color.Black) {
			t.Fatalf("unexpected color from hex, got %v, want %v", col, color.White)
		}
	}
}

func TestCoverConversion(t *testing.T) {

	orig := float32(0.5)
	v := color.FromLinear2sRGB(orig)
	if !math.ApproxEq(color.FromsRGB2Linear(v), orig, math.Epsilon) {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromsRGB2Linear(v), orig)
	}

	if color.FromLinear2sRGB(0) != 0 {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromLinear2sRGB(v), 0)
	}

	if color.FromLinear2sRGB(1) != 1 {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromLinear2sRGB(v), 1)
	}

	if color.FromsRGB2Linear(0) != 0 {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromsRGB2Linear(v), 0)
	}

	if color.FromsRGB2Linear(1) != 1 {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromsRGB2Linear(v), 1)
	}

	color.DisableLut()

	orig = float32(0.5)
	v = color.FromLinear2sRGB(orig)
	if !math.ApproxEq(color.FromsRGB2Linear(v), orig, 1e-6) {
		t.Fatalf("unexpected color conversion, got %v, want %v", color.FromsRGB2Linear(v), orig)
	}
}

func BenchmarkFromHex(b *testing.B) {
	x := "#ffffff"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		color.FromHex(x)
	}
}

func BenchmarkEqual(b *testing.B) {
	c1 := color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}
	c2 := color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}

	for i := 0; i < b.N; i++ {
		color.Equal(c1, c2)
	}
}
