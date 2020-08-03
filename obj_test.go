// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd_test

import (
	"testing"

	"github.com/changkun/ddd"
)

func TestLoadOBJ(t *testing.T) {
	path := "./tests/bunny.obj"
	_, err := ddd.LoadOBJ(path)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}
}

func BenchmarkLoadObj(b *testing.B) {
	path := "./tests/bunny.obj"
	for i := 0; i < b.N; i++ {
		ddd.LoadOBJ(path)
	}
}
