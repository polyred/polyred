// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"image"

	"poly.red/app"
	"poly.red/camera"
	"poly.red/light"
	"poly.red/model"
	"poly.red/render"
	"poly.red/scene"
)

type App struct {
	r   *render.Renderer
	buf *image.RGBA
}

func New() *App {
	cam := camera.NewPerspective()
	// Create a renderer and specify scene and camera
	r := render.NewRenderer(
		render.Scene(scene.NewScene(
			model.StanfordBunny(),
			light.NewPoint(),
		)),
		render.Camera(cam),
	)

	return &App{r: r}
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.buf == nil {
		a.buf = a.r.Render()
	}
	return a.buf, true
}

func main() { app.Run(New()) }
