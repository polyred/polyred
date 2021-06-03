package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
	"time"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
)

func loadMesh(path string) *geometry.TriangleMesh {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	m, err := geometry.LoadOBJ(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	return m
}

func loadTexture(path string) *material.Texture {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load obj model, path: %s, err: %v", path, err))
	}
	var data *image.RGBA
	if v, ok := img.(*image.NRGBA); ok {
		data = (*image.RGBA)(v)
	} else if v, ok := img.(*image.RGBA); ok {
		data = v
	}

	return material.NewTexture(data)
}

func main() {
	result := testing.Benchmark(func(b *testing.B) {
		width, height, msaa := 1920, 1080, 2
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

		l := light.NewPointLight(color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
		s.AddLight(l)

		m := loadMesh("../testdata/bunny.obj")
		tex := loadTexture("../testdata/bunny.png")
		mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
		m.UseMaterial(mat)
		m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
		m.Translate(0, -0, -0.4)
		s.AddMesh(m)

		m = loadMesh("../testdata/ground.obj")
		tex = loadTexture("../testdata/ground.png")
		mat = material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
		m.UseMaterial(mat)
		m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
		m.Translate(0, -0, -0.4)
		s.AddMesh(m)

		r := rend.NewRasterizer(width, height, msaa)
		var buf *image.RGBA
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf = r.Render(s)
		}
		b.StopTimer()

		r.Save(buf, "./render.jpg")
	})

	ns := result.NsPerOp()
	fmt.Printf("BenchmarkRasterizer\t%v\t%v ns/op\t%v fps\n", result.N, ns, 1/(time.Duration(ns)).Seconds())
}
