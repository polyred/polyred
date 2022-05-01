// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package obj_test

import (
	"encoding/json"
	"testing"

	"poly.red/model/obj"
)

func TestParseObj(t *testing.T) {
	path := "../../internal/testdata/bunny.obj"

	f, err := obj.Load(path)
	if err != nil {
		t.Fatal(err)
	}

	_, err = json.Marshal(f)
	if err != nil {
		t.Fatal(err)
	}
}
