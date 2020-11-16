// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd_test

import (
	"fmt"
	"image/color"
	"os"
	"runtime/pprof"
	"testing"

	"changkun.de/x/ddd"
)

var r *ddd.Rasterizer

func newraster() *ddd.Rasterizer {
	width := 800
	height := 500
	r = ddd.NewRasterizer(width, height)
	r.SetCamera(ddd.NewPerspectiveCamera(
		ddd.Vector{-550, 194, 734, 1},
		ddd.Vector{-1000, 0, 0, 1},
		ddd.Vector{0, 1, 1, 0},
		float64(width)/float64(height),
		100,
		600,
		45,
	))

	path := "./tests/bunny.obj"
	m, err := ddd.LoadOBJ(path)
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	m.SetScale(ddd.Vector{1500, 1500, 1500, 0})
	m.SetTranslate(ddd.Vector{-700, -5, 350, 1})
	err = m.SetTexture("./tests/texture.jpg", 150)
	if err != nil {
		panic(fmt.Errorf("cannot load model texture, err: %v", err))
	}
	s := ddd.NewScene()
	s.AddMesh(m)
	s.AddLight(ddd.NewPointLight(color.RGBA{255, 255, 255, 255}, ddd.NewVector(-200, 250, 600, 1), 0.5, 0.6, 1))
	r.SetScene(s)

	return r
}

func TestRasterizer(t *testing.T) {
	r := newraster()

	f, err := os.Create("cpu.pprof")
	if err != nil {
		t.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	for i := 0; i < 1000; i++ {
		r.Render()
	}
	pprof.StopCPUProfile()

	r.Save("./tests/render.jpg")
}

func BenchmarkRasterizer(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			r := newraster()
			r.SetConcurrencySize(int32(block))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Render()
			}
		})
	}
}
