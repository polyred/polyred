// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package alloc_test

import (
	"testing"

	"poly.red/internal/alloc"
)

func TestAlloc(t *testing.T) {
	x := alloc.ID()
	y := alloc.ID()
	if y != x+1 {
		t.Fatalf("expected %v got %v", x+1, y)
	}
}
