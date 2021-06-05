// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package rend_test

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime/pprof"
	"testing"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

func newraster() (*rend.Rasterizer, *rend.Scene) {
	width, height, msaa := 1920, 1080, 2

	c := camera.NewPerspectiveCamera(
		math.NewVector(-0.5, 0.5, 0.5, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		-0.1,
		-3,
	)

	r := rend.NewRasterizer(width, height, msaa)

	path := "../testdata/bunny.obj"
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	m, err := geometry.LoadOBJ(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}

	path = "../testdata/bunny.png"
	f, err = os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}

	tex := material.NewTexture((*image.RGBA)(img.(*image.NRGBA)))
	mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
	m.UseMaterial(mat)

	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Scale(2, 2, 2)
	m.Translate(0, -0, -0.4)

	s := rend.NewScene()
	s.AddMesh(m)
	s.UseCamera(c)

	l := light.NewPointLight(color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
	s.AddLight(l)

	return r, s
}

func TestRasterizer(t *testing.T) {
	r, s := newraster()

	f, err := os.Create("../testdata/cpu.pprof")
	if err != nil {
		t.Fatal(err)
	}
	var buf *image.RGBA
	pprof.StartCPUProfile(f)
	for i := 0; i < 1; i++ {
		buf = r.Render(s)
	}
	pprof.StopCPUProfile()

	utils.Save(buf, "../testdata/render.jpg")
}

func BenchmarkRasterizer(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			r, s := newraster()
			r.SetConcurrencySize(int32(block))

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Render(s)
			}
		})
	}
}
