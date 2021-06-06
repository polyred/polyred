// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package math_test

import (
	"fmt"
	"math/rand"
	"testing"

	"changkun.de/x/ddd/math"
)

func BenchmarkApproxEq(b *testing.B) {
	v1 := 0.000000002
	v2 := 0.000000001

	var bb bool
	for i := 0; i < b.N; i++ {
		bb = math.ApproxEq(v1, v2, math.DefaultEpsilon)
	}
	_ = bb
}

func BenchmarkClamp(b *testing.B) {
	v := 128.0

	var bb float64
	for i := 0; i < b.N; i++ {
		bb = math.Clamp(v, 0, 255)
	}
	_ = bb
}

func BenchmarkClampV(b *testing.B) {
	v := math.Vector{128, 128, 128, 255}

	var bb math.Vector
	for i := 0; i < b.N; i++ {
		bb = math.ClampV(v, 0, 255)
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

	var m math.Matrix
	for i := 0; i < b.N; i++ {
		m = math.ViewportMatrix(w, h)
	}
	_ = m
}
