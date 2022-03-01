// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh_test

import (
	"fmt"
	"math/rand"
	"testing"

	"poly.red/geometry/mesh"
)

func TestLoadOBJ(t *testing.T) {
	path := "../../internal/testdata/bunny.obj"
	_, err := mesh.LoadOBJ(path)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}
}

func BenchmarkLoadOBJ(b *testing.B) {
	path := "../../internal/testdata/bunny-high.obj"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.LoadOBJ(path)
	}
}

func BenchmarkParseFloat(b *testing.B) {
	fs := make([]float64, 100)
	for i := range fs {
		fs[i] = rand.Float64()
	}

	fsstr := make([]string, 100)
	for i := range fsstr {
		fsstr[i] = fmt.Sprintf("%v", fs[i])
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mesh.ParseFloat(fsstr)
	}
}
