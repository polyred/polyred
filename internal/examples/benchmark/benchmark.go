// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"testing"
	"time"

	"poly.red/camera"
	"poly.red/internal/imageutil"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type benchOpts struct {
	width, height, msaa int
	shadowmap           bool
	gammaCorrection     bool
}

func (opt *benchOpts) String() string {
	return fmt.Sprintf("%dx%d-MSAA%dx-ShadowMap-%v-gamma-%v", opt.width, opt.height, opt.msaa, opt.shadowmap, opt.gammaCorrection)
}

func main() {
	opts := []*benchOpts{
		// {960, 540, 1, false, false},
		// {960, 540, 1, true, false},
		// {960, 540, 4, false, false},
		// {960, 540, 4, true, false},
		// {1920, 1080, 1, false, false},
		// {1920, 1080, 1, true, false},
		// {1920, 1080, 4, false, false},
		// {1920, 1080, 4, true, false},
		// {1920 * 2, 1080 * 2, 1, false, false},
		// {1920 * 2, 1080 * 2, 1, true, false},
		// {1920 * 2, 1080 * 2, 4, false, false},
		// {1920 * 2, 1080 * 2, 4, true, false},

		// {960, 540, 1, false, true},
		{960, 540, 1, true, true},
		// {960, 540, 4, false, true},
		// {960, 540, 4, true, true},
		// {1920, 1080, 1, false, true},
		// {1920, 1080, 1, true, true},
		// {1920, 1080, 4, false, true},
		// {1920, 1080, 4, true, true},
		// {1920 * 2, 1080 * 2, 1, false, true},
		// {1920 * 2, 1080 * 2, 1, true, true},
		// {1920 * 2, 1080 * 2, 4, false, true},
		// {1920 * 2, 1080 * 2, 4, true, true},
	}

	for _, opt := range opts {
		for i := 0; i < 1; i++ {
			bench(opt)
		}
	}
}

func bench(opt *benchOpts) {
	result := testing.Benchmark(func(b *testing.B) {
		s := scene.NewScene(light.NewPoint(
			light.Intensity(7),
			light.Color(color.RGBA{0, 0, 0, 255}),
			light.Position(math.NewVec3[float32](4, 4, 2)),
			light.CastShadow(opt.shadowmap),
		), light.NewAmbient(
			light.Intensity(0.5),
			light.Color(color.RGBA{255, 255, 255, 255}),
		))

		m := model.MustLoad("../../testdata/bunny.obj")
		m.Scale(2, 2, 2)
		s.Add(m)

		m = model.MustLoad("../../testdata/ground.obj")
		m.Scale(2, 2, 2)
		s.Add(m)

		r := render.NewRenderer(
			render.Camera(camera.NewPerspective(
				camera.Position(math.NewVec3[float32](0, 0.6, 0.9)),
				camera.ViewFrustum(45, float32(opt.width)/float32(opt.height), 0.1, 2),
			)),
			render.Size(opt.width, opt.height),
			render.MSAA(opt.msaa),
			render.Scene(s),
			render.ShadowMap(opt.shadowmap),
			render.Debug(true),
			render.GammaCorrection(opt.gammaCorrection),
		)

		var buf *image.RGBA
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = r.Render()
		}
		b.StopTimer()
		imageutil.Save(buf, "./benchmark.png")
	})

	ns := time.Duration(result.NsPerOp())
	fmt.Printf("BenchmarkRasterizer-%v\t%v\t%+v ns/op\t%v fps\n", opt, result.N, result.NsPerOp(), 1/(time.Duration(ns)).Seconds())
}
