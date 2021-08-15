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

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/texture"

	"poly.red/internal/utils"
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
		s.Add(light.NewPoint(
			light.Intensity(7),
			light.Color(color.RGBA{0, 0, 0, 255}),
			light.Position(math.NewVec3(4, 4, 2)),
			light.CastShadow(opt.shadowmap),
		), light.NewAmbient(
			light.Intensity(0.5),
			light.Color(color.RGBA{255, 255, 255, 255}),
		))

		m, err := mesh.Load("../../testdata/bunny.obj")
		if err != nil {
			panic(err)
		}
		data := texture.MustLoadImage(
			"../../testdata/bunny.png",
			texture.WithGammaCorrection(opt.gammaCorrection),
		)
		m.SetMaterial(material.NewBlinnPhong(
			material.Texture(texture.NewTexture(
				texture.WithSource(data),
				texture.WithIsotropicMipMap(true),
			)),
			material.Kdiff(0.6), material.Kspec(0.5),
			material.Shininess(150),
			material.ReceiveShadow(opt.shadowmap),
			material.AmbientOcclusion(true),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		m, err = mesh.Load("../../testdata/ground.obj")
		if err != nil {
			panic(err)
		}

		data = texture.MustLoadImage("../../testdata/ground.png",
			texture.WithGammaCorrection(opt.gammaCorrection))
		m.SetMaterial(material.NewBlinnPhong(
			material.Texture(texture.NewTexture(
				texture.WithSource(data),
				texture.WithIsotropicMipMap(true),
			)),
			material.Kdiff(0.6), material.Kspec(0.5),
			material.Shininess(150),
			material.ReceiveShadow(opt.shadowmap),
		))
		m.Scale(2, 2, 2)
		s.Add(m)

		r := render.NewRenderer(
			render.Camera(camera.NewPerspective(
				camera.Position(math.NewVec3(0, 0.6, 0.9)),
				camera.ViewFrustum(45, float64(opt.width)/float64(opt.height), 0.1, 2),
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
		utils.Save(buf, "./benchmark.png")
	})

	ns := time.Duration(result.NsPerOp())
	fmt.Printf("BenchmarkRasterizer-%v\t%v\t%+v ns/op\t%v fps\n", opt, result.N, result.NsPerOp(), 1/(time.Duration(ns)).Seconds())
}
