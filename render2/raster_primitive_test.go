// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"math/rand"
	"testing"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
	render "poly.red/render2"
	"poly.red/shader"
	"poly.red/texture"
	"poly.red/texture/imageutil"
)

var (
	rend *render.Renderer
	m    *mesh.TriangleSoup
	prog shader.Program
)

func init() {
	width, height := 400, 400
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.ViewFrustum(45, float64(width)/float64(height), 0.1, 10),
	)
	rend = render.NewRenderer(render.Size(width, height), render.Camera(cam))

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
}

func TestDrawPrimitives(t *testing.T) {
	buf := rend.NextBuffer()
	rend.DrawPrimitives(buf, m, prog.VertexShader)
	rend.DrawFragments(buf, prog.FragmentShader)
	imageutil.Save(rend.CurrentBuffer().Image(), "../internal/examples/out/draw-primitives.png")
}

func BenchmarkDrawPrimitive(b *testing.B) {
	rand.Seed(42)

	r := render.NewRenderer(render.Size(1920, 1080))
	p := shader.BasicShader{}
	buf := r.NextBuffer()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.DrawPrimitive(buf, p.VertexShader, primitive.NewRandomVertex(), primitive.NewRandomVertex(), primitive.NewRandomVertex())
	}
}

func BenchmarkDrawPrimitives(b *testing.B) {
	rand.Seed(42)

	r := render.NewRenderer(render.Size(1920, 1080))
	p := shader.BasicShader{}
	buf := r.NextBuffer()

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r.DrawPrimitives(buf, m, p.VertexShader)
	}
}
