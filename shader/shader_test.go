// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
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
	"poly.red/render"
	"poly.red/scene"
	"poly.red/shader"
)

func init() {
	rand.Seed(42)
}

func prepare(num int) (*render.Renderer, *buffer.FragmentBuffer, shader.Program, buffer.IndexBuffer, buffer.VertexBuffer) {
	c := camera.NewPerspective(camera.ViewFrustum(50, 1, 0.1, 100))
	r := render.NewRenderer(render.Size(500, 500), render.Camera(c), render.Workers(1))
	buf := buffer.NewBuffer(image.Rect(0, 0, 500, 500))

	m := geometry.NewWith(mesh.NewRandomAs[*mesh.BufferedMesh](num), nil)
	g := scene.NewGroup(m)
	g.Normalize()
	g.TranslateZ(-1)
	return r, buf, &shader.BasicShader{
		ModelMatrix:      g.ModelMatrix(),
		ViewMatrix:       c.ViewMatrix(),
		ProjectionMatrix: c.ProjMatrix(),
	}, m.IndexBuffer(), m.VertexBuffer()
}

func TestShader(t *testing.T) {
	r, buf, prog, idx, tri := prepare(100)
	r.DrawPrimitives(buf, idx, tri, prog.Vertex)
	r.DrawFragments(buf, prog.Fragment, shader.Background(color.White))
	imageutil.Save(buf.Image(), "../internal/examples/out/shader.png")
}

func BenchmarkShaderPrograms(b *testing.B) {
	r, buf, prog, idx, tri := prepare(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawPrimitives(buf, idx, tri, prog.Vertex)
		r.DrawFragments(buf, prog.Fragment)
	}
}
