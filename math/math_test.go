// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"fmt"
	"math/rand"
	"testing"

	"poly.red/math"
)

func TestMinMax(t *testing.T) {
	a, b, c := float32(1.0), float32(2.0), float32(3.0)

	got := math.Min(a, b, c)
	want := float32(1.0)
	if got != want {
		t.Fatalf("unexpected Min, got %v, want %v", got, want)
	}

	got = math.Max(a, b, c)
	want = float32(3.0)
	if got != want {
		t.Fatalf("unexpected Max, got %v, want %v", got, want)
	}
}

func TestRadDeg(t *testing.T) {
	if math.RadToDeg(float32(math.Pi)) != 180 {
		t.Fatalf("unexpected RadToDeg, got %v, want 180.0", math.RadToDeg(float32(math.Pi)))
	}
	if math.DegToRad[float32](180) != float32(math.Pi) {
		t.Fatalf("unexpected DegToRad, got %v, want Pi", math.RadToDeg(float32(math.Pi)))
	}
}

func TestViewportMatrix(t *testing.T) {
	vpMat := math.ViewportMatrix[float32](800, 400)
	want := math.NewMat4[float32](
		400, 0, 0, 400,
		0, 200, 0, 200,
		0, 0, 1, 0,
		0, 0, 0, 1,
	)
	if !vpMat.Eq(want) {
		t.Fatalf("unexpected viewportMatrix, got %v want %v", vpMat, want)
	}
}

func BenchmarkApproxEq(b *testing.B) {
	v1 := float32(0.000000002)
	v2 := float32(0.000000001)

	var bb bool
	for i := 0; i < b.N; i++ {
		bb = math.ApproxEq(v1, v2, math.Epsilon)
	}
	_ = bb
}

func BenchmarkMin(b *testing.B) {
	for j := 1; j < 10000; j *= 2 {
		n := j
		vs := make([]float32, n)
		for i := 0; i < n; i++ {
			vs[i] = rand.Float32()
		}

		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			var v float32
			for i := 0; i < n; i++ {
				v = math.Min(vs...)
			}
			_ = v
		})
	}
}

func BenchmarkMax(b *testing.B) {
	for j := 1; j < 10000; j *= 2 {
		n := j
		vs := make([]float32, n)
		for i := 0; i < n; i++ {
			vs[i] = rand.Float32()
		}

		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			var v float32
			for i := 0; i < n; i++ {
				v = math.Max(vs...)
			}
			_ = v
		})
	}
}

func BenchmarkViewportMatrix(b *testing.B) {
	w, h := float32(1920.0), float32(1080.0)

	var m math.Mat4[float32]
	for i := 0; i < b.N; i++ {
		m = math.ViewportMatrix(w, h)
	}
	_ = m
}
