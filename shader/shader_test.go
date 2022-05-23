// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader_test

import (
	"image"
	"math/rand"
	"testing"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/geometry/mesh"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/shader"
)

func init() {
	rand.Seed(42)
}

func prepare(num int) (*render.Renderer, *buffer.FragmentBuffer, *shader.BasicShader, *scene.Group) {
	c := camera.NewPerspective(camera.ViewFrustum(50, 1, 0.1, 100))
	r := render.NewRenderer(render.Size(500, 500), render.Camera(c), render.Workers(1))
	buf := buffer.NewBuffer(image.Rect(0, 0, 500, 500))

	m := geometry.New(mesh.NewRandomAs[*mesh.BufferedMesh](num))
	g := scene.NewGroup(m)
	g.Normalize()
	g.TranslateZ(-1)
	return r, buf, &shader.BasicShader{
		ModelMatrix:      g.ModelMatrix(),
		ViewMatrix:       c.ViewMatrix(),
		ProjectionMatrix: c.ProjMatrix(),
	}, g
}

func TestShader(t *testing.T) {
	r, buf, prog, g := prepare(100)
	scene.IterObjects(g, func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		prog.ModelMatrix = modelMatrix.MulM(g.ModelMatrix())
		r.DrawPrimitives(buf, g.Triangles(), prog.Vertex)
		return true
	})
	r.DrawFragments(buf, prog.Fragment, shader.Background(color.White))
	imageutil.Save(buf.Image(), "../internal/examples/out/shader.png")
}

func BenchmarkShaderPrograms(b *testing.B) {
	r, buf, prog, g := prepare(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scene.IterObjects(g, func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
			prog.ModelMatrix = modelMatrix.MulM(g.ModelMatrix())
			r.DrawPrimitives(buf, g.Triangles(), prog.Vertex)
			return true
		})
		r.DrawFragments(buf, prog.Fragment)
	}
}
