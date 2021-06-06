// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime/pprof"
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

func main() {
	result := testing.Benchmark(func(b *testing.B) {
		width, height, msaa := 1920, 1080, 2
		// width, height, msaa := 800, 500, 1
		s := rend.NewScene()

		c := camera.NewPerspectiveCamera(
			math.NewVector(-0.5, 0.5, 0.5, 1),
			math.NewVector(0, 0, -0.5, 1),
			math.NewVector(0, 1, 0, 0),
			45,
			float64(width)/float64(height),
			-0.1,
			-3,
		)
		s.UseCamera(c)

		l := light.NewPointLight(20, color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
		s.AddLight(l)

		m := io.MustLoadMesh("../testdata/bunny.obj")
		tex := io.MustLoadTexture("../testdata/bunny.png")
		mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.5, 0.6, 1, 150)
		m.UseMaterial(mat)
		m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
		m.Translate(0, -0, -0.4)
		s.AddMesh(m)

		m = io.MustLoadMesh("../testdata/ground.obj")
		tex = io.MustLoadTexture("../testdata/ground.png")
		mat = material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.5, 0.6, 1, 150)
		m.UseMaterial(mat)
		m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
		m.Translate(0, -0, -0.4)
		s.AddMesh(m)

		r := rend.NewRasterizer(width, height, msaa)

		// cpu pprof
		f, err := os.Create("cpu.pprof")
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)

		var buf *image.RGBA
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = r.Render(s)
		}
		b.StopTimer()
		pprof.StopCPUProfile()
		f.Close()

		utils.Save(buf, "./benchmark.png")
	})

	ns := result.NsPerOp()
	fmt.Printf("BenchmarkRasterizer\t%v\t%v ns/op\t%v fps\n", result.N, ns, 1/(time.Duration(ns)).Seconds())
}
