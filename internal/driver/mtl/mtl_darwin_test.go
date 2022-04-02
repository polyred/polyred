// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mtl_test

import (
	"testing"

	"poly.red/internal/driver/mtl"
)

func TestDevice(t *testing.T) {
	_, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		t.Log(err)
	}
}
