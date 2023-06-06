// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mtl_test

import (
	"testing"

	"poly.red/gpu/mtl"
)

func TestDevice(t *testing.T) {
	_, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		t.Log(err)
	}
}
