// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/internal/gui"
	"poly.red/math"
	render "poly.red/render2"
	"poly.red/shader"
	"poly.red/texture"
	"poly.red/texture/buffer"
	"poly.red/texture/imageutil"
)

func main() {
	// Use a different model
	mod, err := mesh.Load("../../testdata/ground.obj")
	if err != nil {
		panic(err)
	}
	m, ok := mod.(*mesh.TriangleSoup)
	if !ok {
		panic("expect load as an triangle soup")
	}

	m.Normalize()

	tex := texture.NewTexture(
		texture.Image(imageutil.MustLoadImage("../../testdata/uvgrid2.png")),
		texture.IsoMipmap(true),
	)

	width, height := 400, 400
	// camera and renderer
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.ViewFrustum(45, float64(width)/float64(height), 0.1, 10),
	)

	r := render.NewRenderer(
		render.Size(width, height),
		render.Camera(cam),
		render.Workers(2),
	)

	w, err := gui.InitWindow(r, gui.WithTitle("polyred"), gui.WithFPS())
	if err != nil {
		panic(err)
	}

	// Shader
	prog := &shader.TextureShader{
		ModelMatrix: m.ModelMatrix(),
		ViewMatrix:  cam.ViewMatrix(),
		ProjMatrix:  cam.ProjMatrix(),
		Texture:     tex,
	}
	// Handling window resizing
	w.Subscribe(gui.OnResize, func(name gui.EventName, e gui.Event) {
		ev := e.(*gui.SizeEvent)
		cam.SetAspect(float64(ev.Width), float64(ev.Height))
		prog.ViewMatrix = cam.ViewMatrix()
		prog.ModelMatrix = cam.ModelMatrix()
	})

	// Orbit controls
	ctrl := gui.NewOrbitControl(w, cam)
	w.Subscribe(gui.OnMouseUp, ctrl.OnMouse)
	w.Subscribe(gui.OnMouseDown, ctrl.OnMouse)
	w.Subscribe(gui.OnCursor, ctrl.OnCursor)
	w.Subscribe(gui.OnScroll, ctrl.OnScroll)

	// Starts the main rendering loop
	w.MainLoop(func() *buffer.Buffer {
		prog.ModelMatrix = m.ModelMatrix()
		prog.ViewMatrix = cam.ViewMatrix()
		prog.ProjMatrix = cam.ProjMatrix()

		buf := r.NextBuffer()
		r.DrawPrimitives(buf, m, prog.VertexShader)
		r.DrawFragments(buf, prog.FragmentShader)
		return buf
	})
}
