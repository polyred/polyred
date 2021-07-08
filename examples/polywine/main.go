// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"image"
	"runtime"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/gui"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/render"
	"changkun.de/x/polyred/shader"
	"golang.design/x/mainthread"
)

func main() { mainthread.Init(fn) }
func fn() {
	width, height := 400, 400
	gui.InitWindow(
		gui.WithTitle("polyred"),
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
	m := geometry.NewRandomTriangleSoup(1000).(*geometry.BufferedMesh)
	vi, vb := m.GetVertexIndex(), m.GetVertexBuffer()
	gui.Window().Subscribe(gui.OnResize, func(e gui.Event) {
		ev := e.(*gui.SizeEvent)
		cam.SetAspect(float64(ev.Width) / float64(ev.Height))
		prog.ProjectionMatrix = cam.ProjMatrix()
	})
	gui.MainLoop(func(buf *render.Buffer) *image.RGBA {
		cam.RotateX(math.Pi / 100)
		cam.RotateY(math.Pi / 100)
		prog.ModelMatrix = cam.ModelMatrix()

		// 1. Render Primitives
		r.PrimitivePass(buf, prog, vi, vb)

		// 2. Render Screen-space Effects
		r.ScreenPass(buf.Image(), prog.FragmentShader)
		return buf.Image()
	})
}
