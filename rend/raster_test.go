// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package rend_test

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/io"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
	"changkun.de/x/ddd/rend"
	"changkun.de/x/ddd/utils"
)

func newscene() (*rend.Rasterizer, *rend.Scene) {
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

	l := light.NewPointLight(20, color.RGBA{0, 0, 0, 255}, math.NewVector(-200, 250, 600, 1))
	s.AddLight(l)

	m := io.MustLoadMesh("../testdata/bunny.obj")
	tex := io.MustLoadTexture("../testdata/bunny.png")
	mat := material.NewBlinnPhongMaterial(tex, color.RGBA{0, 125, 255, 255}, 0.6, 1, 0.5, 150)
	m.UseMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Scale(2, 2, 2)
	m.Translate(0, -0, -0.4)
	s.AddMesh(m)

	r := rend.NewRasterizer(width, height, msaa)
	return r, s
}

func TestRasterizer(t *testing.T) {
	r, s := newscene()

	f, err := os.Create("cpu.pprof")
	if err != nil {
		t.Fatal(err)
	}
	mem, err := os.Create("mem.pprof")
	if err != nil {
		panic(err)
	}

	var buf *image.RGBA
	pprof.StartCPUProfile(f)
	for i := 0; i < 1; i++ {
		buf = r.Render(s)
	}
	pprof.StopCPUProfile()
	runtime.GC()
	pprof.WriteHeapProfile(mem)
	mem.Close()
	f.Close()

	utils.Save(buf, "../testdata/render.jpg")
}

func BenchmarkRasterizer(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r, s := newscene()
		r.SetConcurrencySize(int32(block))
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Render(s)
			}
		})
	}
}

func BenchmarkForwardPass(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r, s := newscene()
		r.SetConcurrencySize(int32(block))
		r.SetScene(s)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rend.ForwardPass(r)
			}
		})
	}
}

func BenchmarkDeferredPass(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r, s := newscene()
		r.SetConcurrencySize(int32(block))
		r.SetScene(s)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rend.DeferredPass(r)
			}
		})
	}
}

func BenchmarkResetBuf(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r, s := newscene()
		r.SetConcurrencySize(int32(block))
		r.SetScene(s)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rend.ResetBuf(r)
			}
		})
	}
}

func BenchmarkDraw(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r, s := newscene()
		r.SetConcurrencySize(int32(block))
		r.SetScene(s)
		matView := s.Camera.ViewMatrix()
		matProj := s.Camera.ProjMatrix()
		matVP := math.ViewportMatrix(1920, 1080)
		uniforms := map[string]math.Matrix{
			"matModel":  s.Meshes[0].ModelMatrix(),
			"matView":   matView,
			"matProj":   matProj,
			"matVP":     matVP,
			"matNormal": s.Meshes[0].NormalMatrix(),
		}

		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				rend.Draw(r, uniforms, s.Meshes[0].Faces[i%(len(s.Meshes[0].Faces))], s.Meshes[0].Material)
			}
		})
	}
}
