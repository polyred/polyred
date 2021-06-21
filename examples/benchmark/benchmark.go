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
	"changkun.de/x/ddd/scene"
	"changkun.de/x/ddd/utils"
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
		{960, 540, 1, false, false},
		{960, 540, 1, true, false},
		{960, 540, 4, false, false},
		{960, 540, 4, true, false},
		// {1920, 1080, 1, false, false},
		// {1920, 1080, 1, true, false},
		// {1920, 1080, 4, false, false},
		// {1920, 1080, 4, true, false},
		// {1920 * 2, 1080 * 2, 1, false, false},
		// {1920 * 2, 1080 * 2, 1, true, false},
		// {1920 * 2, 1080 * 2, 4, false, false},
		// {1920 * 2, 1080 * 2, 4, true, false},

		{960, 540, 1, false, true},
		{960, 540, 1, true, true},
		{960, 540, 4, false, true},
		{960, 540, 4, true, true},
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
		for i := 0; i < 5; i++ {
			bench(opt)
		}
	}
}

func bench(opt *benchOpts) {
	result := testing.Benchmark(func(b *testing.B) {
		s := scene.NewScene()
		s.SetCamera(camera.NewPerspective(
			math.NewVector(0, 0.6, 0.9, 1),
			math.NewVector(0, 0, 0, 1),
			math.NewVector(0, 1, 0, 0),
			45,
			float64(opt.width)/float64(opt.height),
			0.1,
			2,
		))

		s.Add(light.NewPoint(
			light.WithPointLightIntensity(7),
			light.WithPointLightColor(color.RGBA{0, 0, 0, 255}),
			light.WithPointLightPosition(math.NewVector(4, 4, 2, 1)),
			light.WithShadowMap(opt.shadowmap),
		), light.NewAmbient(
			light.WithAmbientIntensity(0.5),
		))

		m := io.MustLoadMesh("../../testdata/bunny.obj")
		data := io.MustLoadImage("../../testdata/bunny.png")
		m.SetMaterial(material.NewBlinnPhong(
			material.WithBlinnPhongTexture(material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
				material.WithGammaCorrection(opt.gammaCorrection),
			)),
			material.WithBlinnPhongFactors(0.6, 0.5),
			material.WithBlinnPhongShininess(150),
			material.WithBlinnPhongShadow(opt.shadowmap),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		m = io.MustLoadMesh("../../testdata/ground.obj")
		data = io.MustLoadImage("../../testdata/ground.png")
		m.SetMaterial(material.NewBlinnPhong(
			material.WithBlinnPhongTexture(material.NewTexture(
				material.WithImage(data),
				material.WithIsotropicMipMap(true),
				material.WithGammaCorrection(opt.gammaCorrection),
			)),
			material.WithBlinnPhongFactors(0.6, 0.5),
			material.WithBlinnPhongShininess(150),
			material.WithBlinnPhongShadow(opt.shadowmap),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		r := rend.NewRenderer(
			rend.WithSize(opt.width, opt.height),
			rend.WithMSAA(opt.msaa),
			rend.WithScene(s),
			rend.WithShadowMap(opt.shadowmap),
			rend.WithDebug(false),
			rend.WithGammaCorrection(opt.gammaCorrection),
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
	fmt.Printf("BenchmarkRasterizer-%v\t%v\t%+v ns/op\t%v fps\n", opt, result.N, result.NsPerOp(), 1/(time.Duration(ns)).Seconds())
}
