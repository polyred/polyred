// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests_test

import (
	"fmt"
	"testing"

	"poly.red/gpu/tests"
	"poly.red/math"
)

func TestAdd(t *testing.T) {
	if !tests.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := tests.Add(m1, m2)
	sum2 := m1.Add(m2)

	if !sum1.Eq(sum2) {
		t.Fatalf("GPU Add receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v), m1(%v), m2(%v)", sum1, sum2, sum1.Sub(sum2), m1, m2)
	}
}

func TestSub(t *testing.T) {
	if !tests.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := tests.Sub(m1, m2)
	sum2 := m1.Sub(m2)

	if !sum1.Eq(sum2) {
		t.Fatalf("GPU Sub receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v), m1(%v), m2(%v)", sum1, sum2, sum1.Sub(sum2), m1, m2)
	}
}

func TestSqrt(t *testing.T) {
	if !tests.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	r1 := tests.Sqrt(m1)
	r2 := m1.Sqrt()

	if !r1.Eq(r2) {
		t.Fatalf("GPU Sqrt receives different results compare to CPU: GPU(%v)-CPU(%v)=(%v)", r1, r2, r1.Sub(r2))
	}
}

func TestMul(t *testing.T) {
	if !tests.Driver().Available() {
		t.Skip("no GPU device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := tests.Mul(m1, m2)
	sum2 := m1.Mul(m2)

	if !sum1.Eq(sum2) {
		t.Fatalf("GPU Mul receives different results compare to CPU: GPU(%v)*CPU(%v)=(%v), m1(%v), m2(%v)", sum1, sum2, sum1.Sub(sum2), m1, m2)
	}
}

func BenchmarkAdd(b *testing.B) {
	if !tests.Driver().Available() {
		b.Skip("no Metal device available")
	}

	for size := 1; size < 2<<14; size *= 2 {
		m1 := math.NewRandMat[float32](size, size)
		m2 := math.NewRandMat[float32](size, size)

		var outGPU math.Mat[float32]
		var outCPU math.Mat[float32]

		b.Run(fmt.Sprintf("GPU(%vx%v)", size, size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				outGPU = tests.Add(m1, m2)
			}
		})
		b.Run(fmt.Sprintf("CPU(%vx%v)", size, size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				outCPU = m1.Add(m2)
			}
		})

		if !outCPU.Eq(outGPU) {
			b.Fatal("inconsistent results")
		}
	}
}

func BenchmarkMul(b *testing.B) {
	if !tests.Driver().Available() {
		b.Skip("no Metal device available")
	}

	for size := 1 << 5; size < 2<<14; size *= 2 {
		m1 := math.NewRandMat[float32](size, size)
		m2 := math.NewRandMat[float32](size, size)

		var outGPU math.Mat[float32]
		var outCPU math.Mat[float32]

		b.Run(fmt.Sprintf("GPU(%vx%v)", size, size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				outGPU = tests.Mul(m1, m2)
			}
		})
		b.Run(fmt.Sprintf("CPU(%vx%v)", size, size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				outCPU = m1.Mul(m2)
			}
		})

		if !outCPU.Eq(outGPU) {
			b.Fatal("inconsistent results")
		}
	}
}
