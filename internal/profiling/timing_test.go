// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package profiling_test

import (
	"bytes"
	"strings"
	"testing"

	"poly.red/internal/profiling"
)

func TestTimed(t *testing.T) {
	var b bytes.Buffer
	profiling.SetWriter(&b)

	done := profiling.Timed("test")
	done()

	out := b.String()
	if !strings.Contains(out, "test...") {
		t.Fatalf("timed does not print timing info")
	}
}
