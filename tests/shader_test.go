// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package tests

import (
	"image"
	"math/rand"
	"testing"
	"time"

	"poly.red/camera"
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/render"
	"poly.red/shader"
	"poly.red/utils"
)

func init() {
	rand.Seed(time.Now().Unix())
}

func prepare(num int) (*render.Renderer, *render.Buffer, shader.Program, []uint64, []*primitive.Vertex) {
	cam := camera.NewPerspective(
		camera.WithPosition(math.NewVec3(0, 3, 3)),
		camera.WithPerspFrustum(45, 1, 0.1, 10),
	)
	r := render.NewRenderer(
		render.WithSize(500, 500),
		render.WithCamera(cam),
	)
	buf := render.NewBuffer(image.Rect(0, 0, 500, 500))
	prog := &shader.BasicShader{
		ModelMatrix:      math.Mat4I,
		ViewMatrix:       cam.ViewMatrix(),
		ProjectionMatrix: cam.ProjMatrix(),
	}
	idx := make([]uint64, num*3)
	tri := make([]*primitive.Vertex, num*3)
	for i := uint64(0); i < uint64(num*3); i += 3 {
		idx[i] = i
		idx[i+1] = i + 1
		idx[i+2] = i + 2
		tri[i] = &primitive.Vertex{
			Pos: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1),
			UV:  math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, 0, 1),
			Nor: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1).Unit(),
			Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255},
		}
		tri[i+1] = &primitive.Vertex{
			Pos: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1),
			UV:  math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, 0, 1),
			Nor: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1).Unit(),
			Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255},
		}
		tri[i+2] = &primitive.Vertex{
			Pos: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1),
			UV:  math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, 0, 1),
			Nor: math.NewVec4(rand.Float64()*2-1, rand.Float64()*2-1, rand.Float64()*2-1, 1).Unit(),
			Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), 255},
		}
	}

	return r, buf, prog, idx, tri
}

func TestShader(t *testing.T) {
	r, buf, prog, idx, tri := prepare(100)

	// 1. Render Primitives
	r.PrimitivePass(buf, prog, idx, tri)

	// 2. Render Screen-space Effects
	r.ScreenPass(buf.Image(), func(frag primitive.Fragment) color.RGBA {
		if frag.Col == color.Discard {
			return color.White
		}
		return frag.Col
	})

	utils.Save(buf.Image(), "./shader.png")
}

func BenchmarkShaderPrograms(b *testing.B) {
	r, buf, prog, idx, tri := prepare(1000)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.PrimitivePass(buf, prog, idx, tri)
		r.ScreenPass(buf.Image(), func(frag primitive.Fragment) color.RGBA {
			if frag.Col == color.Discard {
				return color.White
			}
			return frag.Col
		})
	}
}
