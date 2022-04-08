// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/internal/imageutil"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/scene"
	"poly.red/scene/object"
	"poly.red/shader"
)

var (
	s *scene.Scene
	c camera.Interface
	r *Renderer
)

func init() {
	w, h, msaa := 800, 600, 2
	s, c = newscene(w, h)
	r = NewRenderer(
		Size(w, h),
		MSAA(msaa),
		Scene(s),
		Background(color.RGBA{0, 127, 255, 255}),
	)
}

func newscene(w, h int) (*scene.Scene, camera.Interface) {
	s := scene.NewScene(light.NewPoint(
		light.Intensity(5),
		light.Color(color.RGBA{0, 0, 0, 255}),
		light.Position(math.NewVec3[float32](-2, 2.5, 6)),
	), light.NewAmbient(
		light.Intensity(0.5),
	))
	m := model.MustLoadAs[*mesh.TriangleMesh]("../internal/testdata/bunny.obj")
	m.Rotate(math.NewVec3[float32](0, 1, 0), -math.Pi/6)
	m.Scale(4, 4, 4)
	m.Translate(0.1, 0, -0.2)
	s.Add(m)
	return s, camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 1.5, 1)),
		camera.LookAt(
			math.NewVec3[float32](0, 0, -0.5),
			math.NewVec3[float32](0, 1, 0),
		),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 3),
	)
}

func TestRender(t *testing.T) {
	w, h, msaa := 1920, 1080, 2
	s, c := newscene(w, h)
	r := NewRenderer(
		Camera(c),
		Size(w, h),
		MSAA(msaa),
		Scene(s),
		Background(color.RGBA{0, 127, 255, 255}),
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

	path := "../internal/testdata/render.png"
	fmt.Printf("render saved at: %s\n", path)
	imageutil.Save(buf, path)
}

func BenchmarkRasterizer(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r.Options(BatchSize(block), Camera(c))
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
		r.Options(BatchSize(block), Camera(c))
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.passForward()
			}
		})
	}
}

func BenchmarkDeferredPass(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		r.Options(BatchSize(block), Camera(c))
		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.passDeferred()
			}
		})
	}
}

func BenchmarkAntiAliasingPass(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.passAntialiasing()
	}
}

func BenchmarkGammaCorrection(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawFragments(r.bufs[0], shader.GammaCorrection)
	}
}

func BenchmarkDrawCall(b *testing.B) {
	for block := 1; block <= 1024; block *= 2 {
		matView := c.ViewMatrix()
		matProj := c.ProjMatrix()
		matVP := math.ViewportMatrix[float32](1920, 1080)

		var (
			m        mesh.Mesh[float32]
			modelMat math.Mat4[float32]
		)
		s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			if o.Type() == object.TypeMesh {
				m = o.(mesh.Mesh[float32])
				modelMat = modelMatrix
				return false
			}
			return true
		})

		mvp := &shader.MVP{
			Model:       modelMat,
			View:        matView,
			ViewInv:     matView.Inv(),
			Proj:        matProj,
			ProjInv:     matProj.Inv(),
			Viewport:    matVP,
			ViewportInv: matVP.Inv(),
			Normal:      modelMat.Inv().T(),
		}

		b.Run(fmt.Sprintf("concurrent-size %d", block), func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			tris := m.Triangles()
			nt := len(tris)
			for i := 0; i < b.N; i++ {
				f := tris[i%int(nt)]
				Draw(r, mvp, f)
			}
		})
	}
}

func BenchmarkInViewport(b *testing.B) {
	v1, v2, v3 := math.NewRandVec4[float32](), math.NewRandVec4[float32](), math.NewRandVec4[float32]()
	for i := 0; i < b.N; i++ {
		r.cullViewFrustum(r.CurrBuffer(), v1, v2, v3)
	}
}
