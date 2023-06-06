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
	"poly.red/geometry"
	"poly.red/internal/imageutil"
	"poly.red/math"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
	"poly.red/shader"
)

type App struct {
	w, h int

	ctrl  *controls.OrbitControl
	r     *render.Renderer
	prog  *shader.TextureShader
	scene *scene.Group
	cam   camera.Interface
	cache *image.RGBA
}

func newApp() *App {
	w, h := 800, 600

	// camera and renderer
	cam := camera.NewPerspective(
		camera.Position(math.NewVec3[float32](0, 3, 3)),
		camera.ViewFrustum(45, float32(w)/float32(h), 0.1, 10),
	)

	r := render.NewRenderer(
		render.Size(w, h),
		render.Camera(cam),
		render.Workers(2),
		render.PixelFormat(buffer.PixelFormatBGRA),
	)
	if runtime.GOOS != "darwin" {
		r.Options(render.PixelFormat(buffer.PixelFormatRGBA))
	}

	s := model.StanfordBunny()
	s.Normalize()
	prog := &shader.TextureShader{
		ViewMatrix: cam.ViewMatrix(),
		ProjMatrix: cam.ProjMatrix(),
		Texture:    buffer.NewTexture(buffer.TextureImage(imageutil.MustLoadImage("../../testdata/bunny.png"))),
	}
	a := &App{w: w, h: h, r: r, prog: prog, cam: cam, scene: s}
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
	a.prog.ViewMatrix = a.cam.ViewMatrix()
	a.prog.ProjMatrix = a.cam.ProjMatrix()

	buf := a.r.NextBuffer()
	scene.IterObjects(a.scene, func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool {
		// Update model matrix before drawing primitives.
		a.prog.ModelMatrix = modelMatrix.MulM(g.ModelMatrix())
		a.r.DrawPrimitives(buf, g.Triangles(), a.prog.Vertex)
		return true
	})
	a.r.DrawFragments(buf, a.prog.Fragment)
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
