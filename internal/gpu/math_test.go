// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gpu_test

import (
	"fmt"
	"testing"

	"poly.red/internal/gpu"
	"poly.red/math"
)

func TestAdd(t *testing.T) {
	if !gpu.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := gpu.Add(m1, m2)
	sum2 := m1.Add(m2)

	if !sum1.Eq(sum2) {
		t.Fatalf("GPU Add receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v), m1(%v), m2(%v)", sum1, sum2, sum1.Sub(sum2), m1, m2)
	}
}

func TestSub(t *testing.T) {
	if !gpu.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := gpu.Sub(m1, m2)
	sum2 := m1.Sub(m2)

	if !sum1.Eq(sum2) {
		t.Fatalf("GPU Sub receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v), m1(%v), m2(%v)", sum1, sum2, sum1.Sub(sum2), m1, m2)
	}
}

func TestSqrt(t *testing.T) {
	if !gpu.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	r1 := gpu.Sqrt(m1)
	r2 := m1.Sqrt()

	if !r1.Eq(r2) {
		t.Fatalf("GPU Sqrt receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v)", r1, r2, r1.Sub(r2))
	}
}

func BenchmarkAdd(b *testing.B) {
	if !gpu.Driver().Available() {
		b.Skip("no Metal device available")
	}

	for size := 1; size < 2<<14; size *= 2 {
		m1 := math.NewRandMat[float32](size, size)
		m2 := math.NewRandMat[float32](size, size)

		var outGPU math.Mat[float32]
		var outCPU math.Mat[float32]

		b.Run(fmt.Sprintf("GPU(%vx%v)", size, size), func(b *testing.B) {
			outGPU = gpu.Add(m1, m2)
		})
		b.Run(fmt.Sprintf("CPU(%vx%v)", size, size), func(b *testing.B) {
			outCPU = m1.Add(m2)
		})

		if !outCPU.Eq(outGPU) {
			b.Fatal("inconsistent results")
		}
	}
}
