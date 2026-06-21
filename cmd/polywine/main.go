// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"
	"log"
	"runtime"

	"poly.red/app"
	"poly.red/app/controls"
	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/light"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type App struct {
	w, h int

	ctrl  *controls.OrbitControl
	r     *render.Renderer
	cam   camera.Interface
	cache *image.RGBA
}

func newApp() *App {
	w, h := 800, 600

	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 10),
	)

	// The Stanford bunny lit by a point light + ambient, rendered through the
	// standard scene renderer (the engine's deferred path), orbit-controllable.
	bunny := model.StanfordBunny()
	bunny.Normalize()
	s := scene.NewScene(
		light.NewPoint(light.Intensity(3), light.Position(math.NewVec3[float32](2, 3, 4))),
		light.NewAmbient(light.Intensity(0.5)),
	)
	s.Add(bunny)

	r := render.NewRenderer(
		render.Size(w, h),
		render.Camera(cam),
		render.Scene(s),
		render.Workers(2),
		render.PixelFormat(buffer.PixelFormatBGRA),
	)
	if runtime.GOOS != "darwin" {
		r.Options(render.PixelFormat(buffer.PixelFormatRGBA))
	}

	a := &App{w: w, h: h, r: r, cam: cam}
	a.ctrl = controls.NewOrbitControl(a.w, a.h, cam)
	return a
}

func (a *App) Size() (int, int) { return a.w, a.h }
func (a *App) OnResize(w, h int) {
	log.Printf("siz:(%vx%v)", w, h)

	a.w = w
	a.h = h
	a.cam.SetAspect(float32(w), float32(h))
	a.r.Options(render.Size(w, h))
	a.cache = nil
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.cache != nil {
		return a.cache, false
	}
	// The orbit control mutates a.cam in place; the renderer holds the same
	// camera and reads it each frame, so a plain Render() reflects the new view.
	a.cache = a.r.Render()
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
