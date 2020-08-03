package main

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/changkun/ddd"
)

func main() {
	result := testing.Benchmark(func(b *testing.B) {
		width, height := 800, 500

		// create rasterizer
		r := ddd.NewRasterizer(width, height)

		// load obj model
		m, err := ddd.LoadOBJ("../tests/bunny.obj")
		if err != nil {
			panic(fmt.Sprintf("cannot load obj model, err: %v", err))
		}

		// set model matrix
		scale := ddd.NewVector(1500, 1500, 1500, 0)
		trans := ddd.NewVector(-700, -5, 350, 1)
		m.SetScale(&scale)
		m.SetTranslate(&trans)

		// set texture
		err = m.SetTexture("../tests/texture.jpg", 150)
		if err != nil {
			panic(fmt.Sprintf("cannot load model texture, err: %v", err))
		}

		// set the camera
		r.SetCamera(ddd.NewPerspectiveCamera(
			ddd.NewVector(-550, 194, 734, 1),
			ddd.NewVector(-1000, 0, 0, 1),
			ddd.NewVector(0, 1, 1, 0),
			float64(width)/float64(height),
			100, 600, 45,
		))
		r.SetCamera(ddd.NewOrthographicCamera(
			ddd.NewVector(-550, 194, 734, 1),
			ddd.NewVector(-1000, 0, 0, 1),
			ddd.NewVector(0, 1, 1, 0),
			-float64(width)/2, float64(width)/2,
			float64(height)/2, -float64(height)/2,
			200, -200,
		))

		// construct scene graph
		s := ddd.NewScene()
		r.SetScene(s)
		s.AddMesh(m)
		l := ddd.NewPointLight(
			color.RGBA{255, 255, 255, 255},
			ddd.NewVector(-200, 250, 600, 1), 0.5, 0.6, 1,
		)
		s.AddLight(l)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Render()
		}
		b.StopTimer()

		r.Save("./render.jpg")
	})

	ns := result.NsPerOp()
	fmt.Printf("render perf: %v fps\n", 1/(time.Duration(ns)).Seconds())
}
