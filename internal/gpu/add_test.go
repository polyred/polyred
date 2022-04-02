// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package gpu_test

import (
	"fmt"
	"reflect"
	"testing"

	"poly.red/internal/gpu"
	"poly.red/math"
)

func TestAdd(t *testing.T) {
	if !gpu.Device.Supported() {
		t.Skip("no Metal device available")
	}

	m1 := math.NewRandMat[float32](10, 10)
	m2 := math.NewRandMat[float32](10, 10)
	sum1 := gpu.Add(m1, m2)
	sum2 := m1.Add(m2)

	if !reflect.DeepEqual(sum1, sum2) {
		t.Fatalf("GPU Add receives different results compare to CPU: GPU(%v) CPU(%v)", sum1, sum2)
	}
}

func BenchmarkAdd(b *testing.B) {
	if !gpu.Device.Supported() {
		b.Skip("no Metal device available")
	}

	for size := 1; size < 2<<10; size *= 2 {
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

		if !reflect.DeepEqual(outCPU, outGPU) {
			b.Fatal("inconsistent results")
		}
	}
}
