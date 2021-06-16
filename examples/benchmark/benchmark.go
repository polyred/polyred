// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

type benchOpts struct {
	width, height, msaa int
	shadowmap           bool
}

func (opt *benchOpts) String() string {
	return fmt.Sprintf("%dx%d-MSAA%dx-ShadowMap-%v", opt.width, opt.height, opt.msaa, opt.shadowmap)
}

func main() {
	opts := []*benchOpts{
		// {960, 540, 1, false},
		// {960, 540, 1, true},
		// {960, 540, 4, false},
		{960, 540, 4, true},
		// {1920, 1080, 1, false},
		// {1920, 1080, 1, true},
		// {1920, 1080, 4, false},
		// {1920, 1080, 4, true},
		// {1920 * 2, 1080 * 2, 1, false},
		// {1920 * 2, 1080 * 2, 1, true},
		// {1920 * 2, 1080 * 2, 4, false},
		// {1920 * 2, 1080 * 2, 4, true},
	}

	for _, opt := range opts {
		bench(opt)
	}
}

func bench(opt *benchOpts) {
	result := testing.Benchmark(func(b *testing.B) {
		s := rend.NewScene()

		c := camera.NewPerspectiveCamera(
			math.NewVector(0, 0.6, 0.9, 1),
			math.NewVector(0, 0, 0, 1),
			math.NewVector(0, 1, 0, 0),
			45,
			float64(opt.width)/float64(opt.height),
			0.1,
			2,
		)
		s.UseCamera(c)

		l := light.NewPointLight(20, color.RGBA{0, 0, 0, 255}, math.NewVector(4, 4, 2, 1))
		s.AddLight(l)

		m := io.MustLoadMesh("../../testdata/bunny.obj")
		tex := io.MustLoadTexture("../../testdata/bunny.png")
		mat := material.NewBlinnPhong(
			material.WithBlinnPhongTexture(tex),
			material.WithBlinnPhongFactors(0.5, 0.6, 1),
			material.WithBlinnPhongShininess(150),
		)
		m.UseMaterial(mat)
		m.Scale(2, 2, 2)
		s.AddMesh(m)

		m = io.MustLoadMesh("../../testdata/ground.obj")
		tex = io.MustLoadTexture("../../testdata/ground.png")
		mat = material.NewBlinnPhong(
			material.WithBlinnPhongTexture(tex),
			material.WithBlinnPhongFactors(0.5, 0.6, 1),
			material.WithBlinnPhongShininess(150),
		)
		m.UseMaterial(mat)
		m.Scale(2, 2, 2)
		s.AddMesh(m)

		r := rend.NewRenderer(
			rend.WithSize(opt.width, opt.height),
			rend.WithMSAA(opt.msaa),
			rend.WithScene(s),
			rend.WithShadowMap(opt.shadowmap),
			rend.WithDebug(true),
		)

		var buf *image.RGBA
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = r.Render()
		}
		b.StopTimer()
		utils.Save(buf, "./benchmark.png")
	})

	ns := time.Duration(result.NsPerOp())
	fmt.Printf("BenchmarkRasterizer-%v\t%v\t%+v/op\t%v fps\n", opt, result.N, ns, 1/(time.Duration(ns)).Seconds())
}
