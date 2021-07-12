// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math_test

import (
	"fmt"
	"math/rand"
	"testing"

	"changkun.de/x/polyred/math"
)

func TestMinMax(t *testing.T) {
	a, b, c := 1.0, 2.0, 3.0

	got := math.Min(a, b, c)
	want := 1.0
	if got != want {
		t.Fatalf("unexpected Min, got %v, want %v", got, want)
	}

	got = math.Max(a, b, c)
	want = 3.0
	if got != want {
		t.Fatalf("unexpected Max, got %v, want %v", got, want)
	}
}

func TestRadDeg(t *testing.T) {
	if math.RadToDeg(math.Pi) != 180 {
		t.Fatalf("unexpected RadToDeg, got %v, want 180.0", math.RadToDeg(math.Pi))
	}
	if math.DegToRad(180) != math.Pi {
		t.Fatalf("unexpected DegToRad, got %v, want Pi", math.RadToDeg(math.Pi))
	}
}

func TestViewportMatrix(t *testing.T) {
	vpMat := math.ViewportMatrix(800, 400)
	want := math.NewMat4(
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
	v1 := 0.000000002
	v2 := 0.000000001

	var bb bool
	for i := 0; i < b.N; i++ {
		bb = math.ApproxEq(v1, v2, math.Epsilon)
	}
	_ = bb
}

func BenchmarkMin(b *testing.B) {
	for j := 1; j < 100000; j *= 2 {
		n := j
		vs := make([]float64, n)
		for i := 0; i < n; i++ {
			vs[i] = rand.Float64()
		}

		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			var v float64
			for i := 0; i < n; i++ {
				v = math.Min(vs...)
			}
			_ = v
		})
	}
}

func BenchmarkMax(b *testing.B) {
	for j := 1; j < 100000; j *= 2 {
		n := j
		vs := make([]float64, n)
		for i := 0; i < n; i++ {
			vs[i] = rand.Float64()
		}

		b.Run(fmt.Sprintf("%d", j), func(b *testing.B) {
			var v float64
			for i := 0; i < n; i++ {
				v = math.Max(vs...)
			}
			_ = v
		})
	}
}

func BenchmarkViewportMatrix(b *testing.B) {
	w, h := 1920.0, 1080.0

	var m math.Mat4
	for i := 0; i < b.N; i++ {
		m = math.ViewportMatrix(w, h)
	}
	_ = m
}
