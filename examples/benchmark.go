// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

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

func main() {
	resolutions := [][]int{
		[]int{800, 500, 1},
		[]int{800, 500, 4},
		[]int{1920, 1080, 1},
		[]int{1920, 1080, 4},
		[]int{1920 * 2, 1080 * 2, 1},
		[]int{1920 * 2, 1080 * 2, 4},
		[]int{1920 * 3, 1080 * 3, 1},
		[]int{1920 * 3, 1080 * 3, 4},
	}

	for _, resolution := range resolutions {
		width, height, msaa := resolution[0], resolution[1], resolution[2]
		result := testing.Benchmark(func(b *testing.B) {
			s := rend.NewScene()

			c := camera.NewPerspectiveCamera(
				math.NewVector(0, 0.6, 0.9, 1),
				math.NewVector(0, 0, 0, 1),
				math.NewVector(0, 1, 0, 0),
				45,
				float64(width)/float64(height),
				-0.1,
				-3,
			)
			s.UseCamera(c)

			l := light.NewPointLight(20, color.RGBA{0, 0, 0, 255}, math.NewVector(4, 4, 2, 1))
			s.AddLight(l)

			m := io.MustLoadMesh("../testdata/bunny.obj")
			tex := io.MustLoadTexture("../testdata/bunny.png")
			mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.5, 0.6, 1, 150)
			m.UseMaterial(mat)
			s.AddMesh(m)

			m = io.MustLoadMesh("../testdata/ground.obj")
			tex = io.MustLoadTexture("../testdata/ground.png")
			mat = material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.5, 0.6, 1, 150)
			m.UseMaterial(mat)
			s.AddMesh(m)

			r := rend.NewRasterizer(width, height, msaa)
			// r.SetDebug(true)

			var buf *image.RGBA
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				buf = r.Render(s)
			}
			b.StopTimer()
			utils.Save(buf, "./benchmark.png")
		})

		ns := result.NsPerOp()
		fmt.Printf("BenchmarkRasterizer-%dx%d-%dxMSAA\t%v\t%v ns/op\t%v fps\n", width, height, msaa, result.N, ns, 1/(time.Duration(ns)).Seconds())
	}
}
