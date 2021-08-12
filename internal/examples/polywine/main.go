// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package main

import (
	"image"
	"image/color"
	"runtime"

	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/internal/gui"
	"poly.red/math"
	"poly.red/render"
	"poly.red/texture"
)

type TextureShader struct {
	ModelMatrix math.Mat4
	ViewMatrix  math.Mat4
	ProjMatrix  math.Mat4
	Texture     *texture.Texture
}

func (s *TextureShader) VertexShader(v primitive.Vertex) primitive.Vertex {
	v.Pos = s.ProjMatrix.MulM(s.ViewMatrix).MulM(s.ModelMatrix).MulV(v.Pos)
	return v
}

func (s *TextureShader) FragmentShader(frag primitive.Fragment) color.RGBA {
	return s.Texture.Query(0, frag.UV.X, 1-frag.UV.Y)
}

func main() {
	width, height := 400, 400
	gui.InitWindow(
		gui.WithTitle("polyred"),
		gui.WithSize(width, height),
		gui.WithFPS(),
	)

	// camera and renderer
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.PerspFrustum(45, float64(width)/float64(height), 0.1, 10),
	)

	r := render.NewRenderer(
		render.WithSize(width, height),
		render.WithCamera(cam),
		render.WithBlendFunc(render.AlphaBlend),
		render.WithThreadLimit(runtime.GOMAXPROCS(1)),
	)

	// Use a different model
	mod, err := mesh.Load("../../testdata/bunny.obj")
	if err != nil {
		panic(err)
	}
	m, ok := mod.(*mesh.TriangleSoup)
	if !ok {
		panic("expect load as an triangle soup")
	}

	m.Normalize()
	vi, vb := m.GetVertexIndex(), m.GetVertexBuffer()

	tex := texture.NewTexture(
		texture.WithSource(texture.MustLoadImage("../../testdata/bunny.png")),
		texture.WithIsotropicMipMap(true),
	)

	// Shader
	prog := &TextureShader{
		ModelMatrix: m.ModelMatrix(),
		ViewMatrix:  cam.ViewMatrix(),
		ProjMatrix:  cam.ProjMatrix(),
		Texture:     tex,
	}
	// Handling window resizing
	gui.Window().Subscribe(gui.OnResize, func(name gui.EventName, e gui.Event) {
		ev := e.(*gui.SizeEvent)
		cam.SetAspect(float64(ev.Width), float64(ev.Height))
		prog.ViewMatrix = cam.ViewMatrix()
		prog.ModelMatrix = cam.ModelMatrix()
	})

	// Orbit controls
	ctrl := gui.NewOrbitControl(cam)
	gui.Window().Subscribe(gui.OnMouseUp, ctrl.OnMouse)
	gui.Window().Subscribe(gui.OnMouseDown, ctrl.OnMouse)
	gui.Window().Subscribe(gui.OnCursor, ctrl.OnCursor)
	gui.Window().Subscribe(gui.OnScroll, ctrl.OnScroll)

	// Starts the main rendering loop
	gui.MainLoop(func(buf *render.Buffer) *image.RGBA {
		prog.ModelMatrix = m.ModelMatrix()
		prog.ViewMatrix = cam.ViewMatrix()
		prog.ProjMatrix = cam.ProjMatrix()

		r.PrimitivePass(buf, prog, vi, vb)
		r.ScreenPass(buf.Image(), prog.FragmentShader)
		return buf.Image()
	})
}
