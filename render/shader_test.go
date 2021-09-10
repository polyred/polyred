// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"image"
	"math/rand"
	"testing"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/render"
	"poly.red/shader"
	"poly.red/texture"
	"poly.red/texture/imageutil"
)

func init() {
	rand.Seed(42)
}

func prepare(num int) (*render.Renderer, *texture.Buffer, shader.Program, []uint64, []*primitive.Vertex) {
	c := camera.NewPerspective(camera.ViewFrustum(50, 1, 0.1, 100))
	r := render.NewRenderer(render.Size(500, 500), render.Camera(c))
	buf := texture.NewBuffer(image.Rect(0, 0, 500, 500))
	prog := &shader.BasicShader{
		ModelMatrix:      math.Mat4I,
		ViewMatrix:       c.ViewMatrix(),
		ProjectionMatrix: c.ProjMatrix(),
	}

	m := mesh.NewRandomTriangleSoup(num).(*mesh.BufferedMesh)
	m.Normalize()
	m.TranslateZ(-1)
	idx := m.GetVertexIndex()
	tri := m.GetVertexBuffer()
	return r, buf, prog, idx, tri
}

func TestShader(t *testing.T) {
	r, buf, prog, idx, tri := prepare(100)

	// 1. Render Primitives
	r.DrawPrimitives(buf, prog.VertexShader, idx, tri)

	// 2. Render Screen-space Effects
	r.DrawFragments(buf, prog.FragmentShader, func(f primitive.Fragment) color.RGBA {
		if f.Col == color.Discard {
			return color.White
		}
		return f.Col
	})

	imageutil.Save(buf.Image(), "../internal/examples/out/shader.png")
}

func BenchmarkShaderPrograms(b *testing.B) {
	r, buf, prog, idx, tri := prepare(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawPrimitives(buf, prog.VertexShader, idx, tri)
		r.DrawFragments(buf, prog.FragmentShader)
	}
}
