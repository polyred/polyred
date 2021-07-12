// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"image/color"
	"runtime"

	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/gui"
	"changkun.de/x/polyred/image"
	"changkun.de/x/polyred/io"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/render"
	"golang.design/x/mainthread"
)

type TextureShader struct {
	ModelMatrix      math.Mat4
	ViewMatrix       math.Mat4
	ProjectionMatrix math.Mat4
	Texture          *image.Texture
}

func (s *TextureShader) VertexShader(v primitive.Vertex) primitive.Vertex {
	v.Pos = s.ProjectionMatrix.MulM(s.ViewMatrix).MulM(s.ModelMatrix).MulV(v.Pos)
	return v
}

func (s *TextureShader) FragmentShader(frag primitive.Fragment) color.RGBA {
	return s.Texture.Query(0, frag.UV.X, 1-frag.UV.Y)
}

func main() { mainthread.Init(fn) }
func fn() {
	width, height := 400, 400
	gui.InitWindow(
		gui.WithTitle("polyred"),
		gui.WithSize(width, height),
		gui.WithFPS(),
	)

	// camera and renderer
	cam := camera.NewPerspective(
		math.NewVec3(0, 3, 3),
		math.NewVec3(0, 0, 0),
		math.NewVec3(0, 1, 0),
		45,
		float64(width)/float64(height),
		0.1, 10,
	)
	r := render.NewRenderer(
		render.WithSize(width, height),
		render.WithCamera(cam),
		render.WithBlendFunc(render.AlphaBlend),
		render.WithThreadLimit(runtime.GOMAXPROCS(0)),
	)

	// Use a different model
	m := io.MustLoadMesh("../../testdata/bunny.obj").(*geometry.TriangleSoup)
	m.Normalize()
	vi, vb := m.GetVertexIndex(), m.GetVertexBuffer()

	tex := image.NewTexture(
		image.WithSource(io.MustLoadImage("../../testdata/bunny.png")),
		image.WithIsotropicMipMap(true),
	)

	// Shader
	prog := &TextureShader{
		ModelMatrix:      math.Mat4I,
		ViewMatrix:       cam.ViewMatrix(),
		ProjectionMatrix: cam.ProjMatrix(),
		Texture:          tex,
	}
	// Handling window resizing
	gui.Window().Subscribe(gui.OnResize, func(name gui.EventName, e gui.Event) {
		ev := e.(*gui.SizeEvent)
		cam.SetAspect(float64(ev.Width), float64(ev.Height))
		prog.ProjectionMatrix = cam.ProjMatrix()
	})

	// Orbit controls
	ctrl := gui.NewOrbitControl(cam)
	gui.Window().Subscribe(gui.OnMouseUp, ctrl.OnMouse)
	gui.Window().Subscribe(gui.OnMouseDown, ctrl.OnMouse)
	gui.Window().Subscribe(gui.OnCursor, ctrl.OnCursor)
	gui.Window().Subscribe(gui.OnScroll, ctrl.OnScroll)

	// Starts the main rendering loop
	gui.MainLoop(func(buf *render.Buffer) *image.RGBA {
		prog.ViewMatrix = cam.ViewMatrix()
		prog.ModelMatrix = cam.ModelMatrix()
		r.PrimitivePass(buf, prog, vi, vb)
		r.ScreenPass(buf.Image(), prog.FragmentShader)
		return buf.Image()
	})
}
