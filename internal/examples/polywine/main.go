// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"
	"log"

	"poly.red/app"
	"poly.red/app/controls"
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/shader"
	"poly.red/texture"
)

type App struct {
	w, h int

	ctrl  *controls.OrbitControl
	r     *render.Renderer
	prog  *shader.TextureShader
	m     *mesh.TriangleSoup
	cam   camera.Interface
	vi    []uint64
	vb    []*primitive.Vertex
	cache *image.RGBA
}

func newApp() *App {
	w, h := 800, 600

	// camera and renderer
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3(0, 3, 3)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 10),
	)

	r := render.NewRenderer(
		render.Size(w, h),
		render.Camera(cam),
		render.Workers(2),
		render.PixelFormat(texture.PixelFormatBGRA),
	)

	m, ok := model.StanfordBunny().(*mesh.TriangleSoup)
	if !ok {
		panic("expect load as an triangle soup")
	}

	m.Normalize()
	vi, vb := m.GetVertexIndex(), m.GetVertexBuffer()
	prog := &shader.TextureShader{
		ModelMatrix: m.ModelMatrix(),
		ViewMatrix:  cam.ViewMatrix(),
		ProjMatrix:  cam.ProjMatrix(),
		Texture:     texture.NewTextureFromImage("../../testdata/bunny.png"),
	}
	a := &App{w: w, h: h, r: r, prog: prog, cam: cam, m: m, vi: vi, vb: vb}
	a.ctrl = controls.NewOrbitControl(a, cam)

	return a
}

func (a *App) Size() (int, int) {
	return a.w, a.h
}

func (a *App) OnResize(w, h int) {
	a.w = w
	a.h = h
	a.cam.SetAspect(float32(w), float32(h))
	a.r.Options(render.Size(w, h))
	a.cache = nil
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.cache != nil {
		return nil, false
	}

	a.prog.ModelMatrix = a.m.ModelMatrix()
	a.prog.ViewMatrix = a.cam.ViewMatrix()
	a.prog.ProjMatrix = a.cam.ProjMatrix()

	buf := a.r.NextBuffer()
	a.r.DrawPrimitives(buf, a.prog.VertexShader, a.vi, a.vb)
	a.r.DrawFragments(buf, a.prog.FragmentShader)
	a.cache = buf.Image()
	return a.cache, true
}

func (a *App) OnMouse(mo app.MouseEvent) {
	log.Println(mo)
	if !a.ctrl.OnMouse(mo) {
		return
	}

	a.cache = nil
}

func (a *App) OnKey(key app.KeyEvent) {
	log.Println(key)
}

func main() {
	app.Run(newApp(),
		app.Title("polywine today"),
		app.MinSize(80, 60),
		app.MaxSize(1920*2, 1080*2),
		app.FPS(false),
	)
}
