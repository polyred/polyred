// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package geometry_test

import (
	"os"
	"testing"

	"changkun.de/x/ddd/geometry"
)

func TestLoadOBJ(t *testing.T) {
	path := "../testdata/bunny.obj"

	f, err := os.Open(path)
	if err != nil {
		t.Errorf("loader: cannot open file %s, err: %v", path, err)
		return
	}
	defer f.Close()

	_, err = geometry.LoadOBJ(f)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}
}

func BenchmarkLoadObj(b *testing.B) {
	path := "../testdata/bunny.obj"
	f, err := os.Open(path)
	if err != nil {
		b.Errorf("loader: cannot open file %s, err: %v", path, err)
		return
	}
	defer f.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		geometry.LoadOBJ(f)
	}
}
