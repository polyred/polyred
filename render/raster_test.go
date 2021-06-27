// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"fmt"
	"image/color"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/light"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/scene"
	"changkun.de/x/polyred/utils"
)

var (
	s *scene.Scene
	r *render.Renderer
)

func init() {
	w, h, msaa := 1920, 1080, 2
	s = newscene(w, h)
	r = render.NewRenderer(
		render.WithSize(w, h),
		render.WithMSAA(msaa),
		render.WithScene(s),
		render.WithBackground(color.RGBA{0, 127, 255, 255}),
	)
}

func newscene(w, h int) *scene.Scene {
	s := scene.NewScene()
	c := camera.NewPerspective(
		math.NewVector(0, 1.5, 1, 1),
		math.NewVector(0, 0, -0.5, 1),
		math.NewVector(0, 1, 0, 0),
		45,
		float64(w)/float64(h),
		0.1,
		3,
	)
	s.SetCamera(c)

	s.Add(light.NewPoint(
		light.WithPointLightIntensity(5),
		light.WithPointLightColor(color.RGBA{0, 0, 0, 255}),
		light.WithPointLightPosition(math.NewVector(-2, 2.5, 6, 1)),
	), light.NewAmbient(
		light.WithAmbientIntensity(0.5),
	))

	m := io.MustLoadMesh("../testdata/bunny.obj")
	data := io.MustLoadImage("../testdata/bunny.png")
	mat := material.NewBlinnPhong(
		material.WithBlinnPhongTexture(image.NewTexture(
			image.WithSource(data),
			image.WithIsotropicMipMap(true),
		)),
		material.WithBlinnPhongFactors(0.8, 1),
		material.WithBlinnPhongShininess(100),
	)
	m.SetMaterial(mat)
	m.Rotate(math.NewVector(0, 1, 0, 0), -math.Pi/6)
	m.Scale(4, 4, 4)
	m.Translate(0.1, 0, -0.2)
	s.Add(m)
	return s
}

func TestRasterizer(t *testing.T) {
	w, h, msaa := 1920, 1080, 2
	s := newscene(w, h)
	r := render.NewRenderer(
		render.WithSize(w, h),
		render.WithMSAA(msaa),
		render.WithScene(s),
		render.WithBackground(color.RGBA{0, 127, 255, 255}),
	)

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
		r.UpdateOptions(
			render.WithConcurrency(int32(block)),
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
		r.UpdateOptions(
			render.WithConcurrency(int32(block)),
		)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				render.PassForward(r)
			}
		})
	}
}

func BenchmarkDeferredPass(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r.UpdateOptions(
			render.WithConcurrency(int32(block)),
		)
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				render.PassDeferred(r)
			}
		})
	}
}

func BenchmarkAntiAliasingPass(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		render.PassAntiAliasing(r)
	}
}

func BenchmarkResetBuf(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				render.ResetGBuf(r)
				render.ResetFrameBuf(r)
			}
		})
	}
}

func BenchmarkDraw(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		matView := s.GetCamera().ViewMatrix()
		matProj := s.GetCamera().ProjMatrix()
		matVP := math.ViewportMatrix(1920, 1080)

		var (
			mesh     geometry.Mesh
			modelMat math.Matrix
		)
		s.IterObjects(func(o object.Object, modelMatrix math.Matrix) bool {
			if o.Type() == object.TypeMesh {
				mesh = o.(geometry.Mesh)
				modelMat = modelMatrix
				return false
			}
			return true
		})

		uniforms := map[string]interface{}{
			"matModel":  modelMat,
			"matView":   matView,
			"matProj":   matProj,
			"matVP":     matVP,
			"matNormal": modelMat.Inv().T(),
		}

		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			var (
				ts  = []*primitive.Triangle{}
				mat material.Material
				nt  = mesh.NumTriangles()
			)

			mesh.Faces(func(f primitive.Face, m material.Material) bool {
				mat = m
				f.Triangles(func(t *primitive.Triangle) bool {
					ts = append(ts, t)
					return true
				})
				return true
			})

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				f := ts[i%int(nt)]
				render.Draw(r, uniforms, f, modelMat, mat)
			}
		})
	}
}
