// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"

	"poly.red/app"
	"poly.red/app/controls"
	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type App struct {
	w, h int

	ctrl  *controls.OrbitControl
	r     *render.Renderer
	c     camera.Interface
	s     *scene.Scene
	cache *image.RGBA
}

func newApp(objPath string) *App {
	w, h := 800, 600

	// camera and renderer
	c := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 10),
	)

	s := scene.NewScene(model.MustLoad(objPath))
	s.Normalize()

	r := render.NewRenderer(
		render.Size(w, h),
		render.Camera(c),
		render.Workers(2),
		render.PixelFormat(buffer.PixelFormatBGRA),
		render.Scene(s),
	)
	a := &App{w: w, h: h, r: r, c: c, s: s}
	a.ctrl = controls.NewOrbitControl(a, c)

	return a
}

func (a *App) Size() (int, int) {
	return a.w, a.h
}

func (a *App) OnResize(w, h int) {
	a.w = w
	a.h = h
	a.c.SetAspect(float32(w), float32(h))
	a.r.Options(render.Size(w, h))
	a.cache = nil
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.cache != nil {
		return nil, false
	}

	// FIXME: render does not work yet.
	// Blinn-Phong shader have issue to compute the actual color.
	a.cache = a.r.Render()
	return a.cache, true
}

func (a *App) OnMouse(mo app.MouseEvent) {
	if !a.ctrl.OnMouse(mo) {
		return
	}

	a.cache = nil
}
