// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import (
	"math"
	"testing"
)

// TestSRGB checks the author-once SRGB kernel run as Go: the linear-to-sRGB
// transfer in both regimes around the 0.0031308 knee.
func TestSRGB(t *testing.T) {
	srgb := func(v float64) float32 {
		if v <= 0.0031308 {
			return float32(v * 12.92)
		}
		return float32(1.055*math.Pow(v, 0.4166666666) - 0.055)
	}
	for _, v := range []float64{0, 0.0031308, 0.1, 0.5, 1} {
		in := []float32{float32(v)}
		out := make([]float32, 1)
		SRGB(0, in, out)
		if d := out[0] - srgb(v); d > 1e-6 || d < -1e-6 {
			t.Errorf("SRGB(%v) = %v, want %v", v, out[0], srgb(v))
		}
	}
}
