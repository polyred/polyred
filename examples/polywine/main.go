// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"math/rand"
	"runtime"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/gui"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/shader"
	"golang.design/x/mainthread"
)

func main() { mainthread.Init(fn) }
func fn() {
	width, height := 400, 400
	w := gui.NewWindow(
		gui.WithTitle(fmt.Sprintf("polyred - %dx%d", width, height)),
		gui.WithSize(width, height),
		gui.WithFPS(),
	)
	cam := camera.NewPerspective(
		math.NewVec4(0, 3, 3, 1),
		math.NewVec4(0, 0, 0, 1),
		math.NewVec4(0, 1, 0, 0),
		45,
		float64(width)/float64(height),
		0.1, 10,
	)
	r := render.NewRenderer(
		render.WithSize(width, height),
		render.WithCamera(cam),
		render.WithThreadLimit(runtime.GOMAXPROCS(0)),
	)
	prog := &shader.BasicShader{
		ModelMatrix:      math.Mat4I,
		ViewMatrix:       cam.ViewMatrix(),
		ProjectionMatrix: cam.ProjMatrix(),
	}
	buf := render.NewBuffer(image.Rect(0, 0, width, height))
	idx, tri := geo(1000)

	w.MainLoop(func() *image.RGBA {
		buf.Clear()
		cam.RotateX(math.Pi / 100)
		cam.RotateY(math.Pi / 100)
		prog.ModelMatrix = cam.ModelMatrix()

		// 1. Render Primitives
		r.PrimitivePass(buf, prog, idx, tri)

		// 2. Render Screen-space Effects
		r.ScreenPass(buf.Image(), func(frag primitive.Fragment) color.RGBA {
			if frag.Col == color.Discard {
				return color.Black
			}
			return frag.Col
		})
		return buf.Image()
	})
}

func geo(num int) ([]uint64, []*primitive.Vertex) {
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
	return idx, tri
}
