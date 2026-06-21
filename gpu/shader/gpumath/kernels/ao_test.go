// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package kernels

import "testing"

// TestAO checks the author-once AO kernel run as Go: a non-AO fragment is
// untouched, and an AO fragment is only ever darkened (the pow(.,10000) factor is
// in [0,1]). The factor is too float-sensitive for an exact assertion (the
// engine's GPU/CPU AO parity uses a tolerance for the same reason).
func TestAO(t *testing.T) {
	orig := []float32{200, 150, 100}

	// aoflag = 0: fragment skips AO, color unchanged.
	color := []float32{200, 150, 100, 255}
	AO(0, []float32{4, 4, 0, 0}, []float32{0}, make([]float32, 64), color, []float32{8, 8, 0, 0})
	for i := 0; i < 3; i++ {
		if color[i] != orig[i] {
			t.Errorf("aoflag=0 chan %d changed: got %v want %v", i, color[i], orig[i])
		}
	}

	// aoflag = 1: AO only darkens, so each channel ends up in [0, orig].
	color2 := []float32{200, 150, 100, 255}
	AO(0, []float32{4, 4, 0, 0}, []float32{1}, make([]float32, 64), color2, []float32{8, 8, 0, 0})
	for i := 0; i < 3; i++ {
		if color2[i] < 0 || color2[i] > orig[i] {
			t.Errorf("AO chan %d = %v, want in [0, %v]", i, color2[i], orig[i])
		}
	}
}
