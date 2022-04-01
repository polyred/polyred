// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package imageutil_test

import (
	"testing"

	"poly.red/internal/imageutil"
)

func TestDiff(t *testing.T) {
	img := imageutil.MustLoadImage("../../internal/testdata/bunny.png")
	_, error := imageutil.Diff(img, img, imageutil.MseKernel)
	if error != 0 {
		t.Fatalf("same image returns a non-zero diff error")
	}
}
