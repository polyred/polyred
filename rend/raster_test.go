// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

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

func newscene() *rend.Renderer {
	width, height, msaa := 1920, 1080, 2

	s := rend.NewScene()
	c := camera.NewPerspectiveCamera(
		math.NewVector(0, 1.5, 1, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		0.1,
		3,
	)
	s.UseCamera(c)

	l := light.NewPointLight(20, color.RGBA{0, 0, 0, 255}, math.NewVector(-2, 2.5, 6, 1))
	s.AddLight(l)

	m := io.MustLoadMesh("../testdata/bunny.obj")
	tex := io.MustLoadTexture("../testdata/bunny.png")
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(tex),
		material.WithBlinnPhongFactors(0.5, 0.6, 1),
		material.WithBlinnPhongShininess(150),
	)
	m.UseMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Scale(4, 4, 4)
	m.Translate(0.1, 0, -0.2)
	s.AddMesh(m)

	r := rend.NewRenderer(
		rend.WithSize(width, height),
		rend.WithMSAA(msaa),
		rend.WithScene(s),
		rend.WithBackground(color.RGBA{0, 127, 255, 255}),
	)
	return r
}

func TestRasterizer(t *testing.T) {
	r := newscene()

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
	for i := 0; i < 10; i++ {
		buf = r.Render()
	}
	pprof.StopCPUProfile()
	runtime.GC()
	pprof.WriteHeapProfile(mem)
	mem.Close()
	f.Close()

	path := "../testdata/render.jpg"
	fmt.Printf("render saved at: %s\n", path)
	utils.Save(buf, path)
}

func BenchmarkRasterizer(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r := newscene()
		r.UpdateOptions(
			rend.WithConcurrency(int32(block)),
		)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.Render()
			}
		})
	}
}

func BenchmarkForwardPass(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r := newscene()
		r.UpdateOptions(
			rend.WithConcurrency(int32(block)),
		)
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
		r := newscene()
		r.UpdateOptions(
			rend.WithConcurrency(int32(block)),
		)
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
		r := newscene()
		r.UpdateOptions(
			rend.WithConcurrency(int32(block)),
		)
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
		r := newscene()
		r.UpdateOptions(
			rend.WithConcurrency(int32(block)),
		)
		matView := r.GetScene().Camera.ViewMatrix()
		matProj := r.GetScene().Camera.ProjMatrix()
		matVP := math.ViewportMatrix(1920, 1080)
		uniforms := map[string]interface{}{
			"matModel":  r.GetScene().Meshes[0].ModelMatrix(),
			"matView":   matView,
			"matProj":   matProj,
			"matVP":     matVP,
			"matNormal": r.GetScene().Meshes[0].NormalMatrix(),
		}

		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			mat := r.GetScene().Meshes[0].Material
			for i := 0; i < b.N; i++ {
				f := r.GetScene().Meshes[0].Faces[i%(len(r.GetScene().Meshes[0].Faces))]
				rend.Draw(r, uniforms, f, mat)
			}
		})
	}
}
