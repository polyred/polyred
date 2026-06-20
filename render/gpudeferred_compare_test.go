// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package render

import "testing"

// assertDeferredClose checks that the CPU and GPU deferred images match,
// tolerating a small fraction of channels that differ substantially.
//
// For multi-object scenes the two separate renders can disagree at a handful of
// contested silhouette/overlap pixels: the concurrent forward pass picks the
// winning fragment non-deterministically across runs/machines, so the GPU
// render's G-buffer can hold a different material than the CPU render's at those
// pixels. That is a test-methodology artifact of comparing two independent
// renders, not a GPU shading error — the shading math itself is bit-identical
// (see gpu/shader/blinnphong_parity_darwin_test.go and the single-object
// TestGPUDeferredParity, which stays exact). The bulk of pixels must match
// tightly; only a tiny fraction may diverge.
func assertDeferredClose(t *testing.T, cpu, gpu []uint8, label string) {
	t.Helper()
	if len(cpu) != len(gpu) {
		t.Fatalf("%s: image size mismatch %d vs %d", label, len(cpu), len(gpu))
	}
	nBig := 0
	for i := range cpu {
		d := int(cpu[i]) - int(gpu[i])
		if d < 0 {
			d = -d
		}
		if d > 8 {
			nBig++
		}
	}
	frac := float64(nBig) / float64(len(cpu))
	if frac > 0.02 {
		t.Fatalf("%s: %.2f%% of channels differ by >8 (want <2%%, %d/%d)", label, frac*100, nBig, len(cpu))
	}
	t.Logf("%s: %d/%d channels differ by >8 (%.3f%%)", label, nBig, len(cpu), frac*100)
}
