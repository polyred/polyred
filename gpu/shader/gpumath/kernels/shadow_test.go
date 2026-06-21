// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import "testing"

// TestShadow checks the author-once Shadow kernel run as Go: a non-receiving
// fragment is untouched, and a receiving fragment with no shadow maps keeps its
// color (factor 1) after the round/clamp/floor quantization.
func TestShadow(t *testing.T) {
	// recv = 0: fragment does not receive shadow, color unchanged.
	color := []float32{200, 150, 100, 255}
	Shadow(0, []float32{0, 0, 0, 0}, []float32{0}, []float32{0}, []float32{0}, color, []float32{4, 0, 0, 0})
	if color[0] != 200 || color[1] != 150 || color[2] != 100 {
		t.Errorf("recv=0 should leave color unchanged, got %v", color[:3])
	}

	// recv = 1, n = 0 maps: occ = 0, factor = pow(0.5,0) = 1; color is quantized
	// via floor(clamp(round(c),0,255)).
	color2 := []float32{200.4, 149.6, 100, 255}
	su := []float32{4, 0, 0, 0} // width=4, depthLen=0, n=0
	Shadow(0, []float32{0, 0, 0, 0}, []float32{1}, []float32{0}, []float32{0}, color2, su)
	want := []float32{200, 150, 100}
	for i := 0; i < 3; i++ {
		if color2[i] != want[i] {
			t.Errorf("recv=1 n=0 chan %d = %v, want %v", i, color2[i], want[i])
		}
	}
}
