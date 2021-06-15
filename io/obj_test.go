// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package io_test

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"changkun.de/x/ddd/io"
)

func TestLoadOBJ(t *testing.T) {
	path := "../testdata/bunny.obj"

	f, err := os.Open(path)
	if err != nil {
		t.Errorf("loader: cannot open file %s, err: %v", path, err)
		return
	}
	defer f.Close()

	_, err = io.LoadOBJ(f)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}
}

func BenchmarkLoadOBJ(b *testing.B) {
	path := "../testdata/bunny-high.obj"
	f, err := os.Open(path)
	if err != nil {
		b.Errorf("loader: cannot open file %s, err: %v", path, err)
		return
	}
	defer f.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.LoadOBJ(f)
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
		io.ParseFloat(fsstr)
	}

}
