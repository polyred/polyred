// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package main

import (
	"flag"
	"image"
	"log"

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

// New builds the renderer for the demo. By default the renderer is GPU-by-default:
// on macOS it offloads the deferred shading pass to Metal (the forward rasterizer
// still runs on the CPU), elsewhere it runs all-CPU. Pass forceCPU to run entirely
// on the CPU.
func New(forceCPU bool) *App {
	cam := camera.NewPerspective()
	opts := []render.Option{
		render.Scene(scene.NewScene(
			model.StanfordBunny(),
			light.NewPoint(),
		)),
		render.Camera(cam),
	}
	if forceCPU {
		opts = append(opts, render.CPU())
	}
	return &App{r: render.NewRenderer(opts...)}
}

func (a *App) Draw() (*image.RGBA, bool) {
	if a.buf == nil {
		a.buf = a.r.Render()
	}
	return a.buf, true
}

func main() {
	cpu := flag.Bool("cpu", false, "force the CPU rasterizer instead of the default GPU-offloaded renderer")
	flag.Parse()
	if *cpu {
		log.Println("simple: rendering on the CPU")
	} else {
		log.Println("simple: rendering with the default renderer (GPU deferred shading where available; -cpu to force CPU)")
	}
	app.Run(New(*cpu))
}
