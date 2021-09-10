// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"image"
	"math/rand"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/render"
	"poly.red/shader"
	"poly.red/texture"
	"poly.red/texture/imageutil"
)

var (
	rend *render.Renderer
	m    *mesh.TriangleSoup
	prog shader.Program
	buf  *texture.Buffer
)

func init() {
	width, height := 400, 400
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.ViewFrustum(45, float32(width)/float32(height), 0.1, 10),
	)
	rend = render.NewRenderer(
		render.Size(width, height),
		render.Camera(cam),
		render.Blending(render.AlphaBlend),
	)

	// Use a different model
	mod, err := mesh.Load("../internal/testdata/bunny.obj")
	if err != nil {
		panic(err)
	}
	var ok bool
	m, ok = mod.(*mesh.TriangleSoup)
	if !ok {
		panic("expect load as an triangle soup")
	}
	m.Normalize()

	tex := texture.NewTexture(
		texture.Image(imageutil.MustLoadImage("../internal/testdata/bunny.png")),
		texture.IsoMipmap(true),
	)
	prog = &shader.TextureShader{
		ModelMatrix: m.ModelMatrix(),
		ViewMatrix:  cam.ViewMatrix(),
		ProjMatrix:  cam.ProjMatrix(),
		Texture:     tex,
	}
	buf = texture.NewBuffer(image.Rect(0, 0, width, height))
}

func TestDrawPrimitives(t *testing.T) {
	vi, vb := m.GetVertexIndex(), m.GetVertexBuffer()
	rend.DrawPrimitives(buf, prog.VertexShader, vi, vb)
	rend.DrawFragments(buf, prog.FragmentShader)
	imageutil.Save(buf.Image(), "../internal/examples/out/draw-primitives.png")
}

func BenchmarkDrawPrimitive(b *testing.B) {
	rand.Seed(42)

	buf := texture.NewBuffer(image.Rect(0, 0, 1920, 1080))
	r := render.NewRenderer()
	p := shader.BasicShader{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawPrimitive(buf, p.VertexShader, &primitive.Triangle{})
	}
}
