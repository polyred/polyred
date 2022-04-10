// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package model_test

import (
	"fmt"
	"log"
	"os"
	"testing"

	"poly.red/geometry/mesh"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
)

func TestLoadOBJ(t *testing.T) {
	path := "../internal/testdata/bunny.obj"
	g, err := model.Load(path)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}

	scene.IterObjects(g, func(m *mesh.TriangleMesh, modelMatrix math.Mat4[float32]) bool {
		t.Log(m, modelMatrix)
		return true
	})
}

func TestLoadSponza(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("cannot get home dir: %v", err)
	}

	path := fmt.Sprintf("%s/Dropbox/Data/sponza.obj", home)
	g, err := model.Load(path)
	if err != nil {
		t.Fatalf("cannot load obj model, path: %s, err: %v", path, err)
	}

	scene.IterObjects(g, func(m *mesh.TriangleMesh, modelMatrix math.Mat4[float32]) bool {
		t.Log(m, modelMatrix)
		return true
	})
}

func BenchmarkLoadOBJ(b *testing.B) {
	path := "../internal/testdata/bunny.obj"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model.MustLoad(path)
	}
}
