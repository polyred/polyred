// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render_test

import (
	"image"
	"math/rand"
	"testing"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/shader"
)

var (
	rend *render.Renderer
	s    *scene.Group
	prog *shader.TextureShader
	buf  *buffer.FragmentBuffer
)

func init() {
	width, height := 400, 400
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
		camera.ViewFrustum(45, float32(width)/float32(height), 0.1, 10),
	)
	rend = render.NewRenderer(
		render.Size(width, height),
		render.Camera(cam),
		render.Blending(render.AlphaBlend),
	)

	// Use a different model
	s = model.MustLoad("../internal/testdata/bunny.obj")
	s.Normalize()

	tex := buffer.NewTexture(
		buffer.TextureImage(imageutil.MustLoadImage("../internal/testdata/bunny.png")),
		buffer.TextureIsoMipmap(true),
	)
	prog = &shader.TextureShader{
		ViewMatrix: cam.ViewMatrix(),
		ProjMatrix: cam.ProjMatrix(),
		Texture:    tex,
	}
	buf = buffer.NewBuffer(image.Rect(0, 0, width, height))
}

func TestDrawPrimitives(t *testing.T) {
	scene.IterObjects(s, func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		prog.ModelMatrix = modelMatrix.MulM(g.ModelMatrix())
		rend.DrawPrimitives(buf, g.Triangles(), prog.Vertex)
		return true
	})
	rend.DrawFragments(buf, prog.Fragment)
	imageutil.Save(buf.Image(), "../internal/examples/out/draw-primitives.png")
}

func BenchmarkDrawPrimitive(b *testing.B) {
	rand.Seed(42)

	buf := buffer.NewBuffer(image.Rect(0, 0, 1920, 1080))
	r := render.NewRenderer()
	p := shader.BasicShader{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		render.DrawPrimitive(r, buf, &primitive.Triangle{
			V1: &primitive.Vertex{},
			V2: &primitive.Vertex{},
			V3: &primitive.Vertex{},
		}, p.Vertex)
	}
}
