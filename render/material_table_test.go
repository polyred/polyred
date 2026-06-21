// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"testing"

	"poly.red/material"
)

// TestMatAt pins the per-frame flat material resolution, including the negative
// index "use vertex color" fallback that replaced the old negative material ID.
// This is the resolution the de-globalization moved from the material pool into
// the renderer; a registry/index bug must not silently drop it.
func TestMatAt(t *testing.T) {
	a := material.NewBlinnPhong()
	b := material.NewBlinnPhong()
	table := []*material.BlinnPhong{a, b}

	if matAt(table, -1) != nil {
		t.Error("matAt(-1) should be nil (use vertex color)")
	}
	if matAt(table, 0) != a || matAt(table, 1) != b {
		t.Error("matAt should return the material at the flat index")
	}
	if matAt(table, 2) != nil || matAt(table, 1<<30) != nil {
		t.Error("matAt(out-of-range) should be nil")
	}
	if matAt(nil, 0) != nil {
		t.Error("matAt(empty, 0) should be nil")
	}
}
