// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/image"
	"poly.red/io"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/utils"
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
		s := scene.NewScene()
		s.SetCamera(camera.NewPerspective(
			camera.WithPosition(math.NewVec3(0, 0.6, 0.9)),
			camera.WithPerspFrustum(
				45, float64(opt.width)/float64(opt.height), 0.1, 2),
		))

		s.Add(light.NewPoint(
			light.WithPointLightIntensity(7),
			light.WithPointLightColor(color.RGBA{0, 0, 0, 255}),
			light.WithPointLightPosition(math.NewVec3(4, 4, 2)),
			light.WithPointLightShadowMap(opt.shadowmap),
		), light.NewAmbient(
			light.WithAmbientIntensity(0.5),
			light.WithAmbientColor(color.RGBA{255, 255, 255, 255}),
		))

		m, err := mesh.Load("../../testdata/bunny.obj")
		if err != nil {
			panic(err)
		}
		data := io.MustLoadImage(
			"../../testdata/bunny.png",
			io.WithGammaCorrection(opt.gammaCorrection),
		)
		m.SetMaterial(material.NewBlinnPhong(
			material.WithBlinnPhongTexture(image.NewTexture(
				image.WithSource(data),
				image.WithIsotropicMipMap(true),
			)),
			material.WithBlinnPhongFactors(0.6, 0.5),
			material.WithBlinnPhongShininess(150),
			material.WithBlinnPhongShadow(opt.shadowmap),
			material.WithBlinnPhongAmbientOcclusion(true),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		m, err = mesh.Load("../../testdata/ground.obj")
		if err != nil {
			panic(err)
		}

		data = io.MustLoadImage("../../testdata/ground.png",
			io.WithGammaCorrection(opt.gammaCorrection))
		m.SetMaterial(material.NewBlinnPhong(
			material.WithBlinnPhongTexture(image.NewTexture(
				image.WithSource(data),
				image.WithIsotropicMipMap(true),
			)),
			material.WithBlinnPhongFactors(0.6, 0.5),
			material.WithBlinnPhongShininess(150),
			material.WithBlinnPhongShadow(opt.shadowmap),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		r := render.NewRenderer(
			render.WithSize(opt.width, opt.height),
			render.WithMSAA(opt.msaa),
			render.WithScene(s),
			render.WithShadowMap(opt.shadowmap),
			render.WithDebug(true),
			render.WithGammaCorrection(opt.gammaCorrection),
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
