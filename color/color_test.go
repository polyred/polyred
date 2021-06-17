// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package color_test

import (
	"math/rand"
	"testing"

	"changkun.de/x/ddd/color"
)

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
