// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"

	"poly.red/app"
	"poly.red/camera"
	"poly.red/geometry/mesh"
	"poly.red/light"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type App struct {
	c     camera.Interface
	s     *scene.Scene
	r     *render.Renderer
	w, h  int
	image *image.RGBA
}

func New() *App {
	w, h := 800, 600

	// Create a scene graph
	s := scene.NewScene()

	// Create and add a point light and a bunny to the scene graph
	s.Add(light.NewPoint(), model.StanfordBunnyAs[*mesh.TriangleMesh]())

	// Create a camera for the rendering
	c := camera.NewPerspective()

	// Create a renderer and specify scene and camera
	r := render.NewRenderer(render.Size(w, h), render.Scene(s), render.Camera(c))

	return &App{c: c, s: s, r: r, w: w, h: h}
}

func (a *App) Size() (int, int) {
	return a.w, a.h
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.image == nil {
		a.image = a.r.Render()
	}
	return a.image, true
}

func main() { app.Run(New()) }
