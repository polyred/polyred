// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/camera"
	"poly.red/scene"
	"poly.red/texture/buffer"
)

// Opt represents a rendering Opt
type Opt func(r *Renderer)

func Size(width, height int) Opt {
	return func(r *Renderer) {
		r.width = width
		r.height = height
	}
}

func Camera(cam camera.Interface) Opt {
	return func(r *Renderer) {
		r.renderCamera = cam
		if _, ok := cam.(*camera.Perspective); ok {
			r.renderPerspect = true
		}
	}
}

func Scene(s *scene.Scene) Opt {
	return func(r *Renderer) {
		r.scene = s
	}
}

func Background(c color.RGBA) Opt {
	return func(r *Renderer) {
		r.background = c
	}
}

func MSAA(n int) Opt {
	return func(r *Renderer) {
		r.msaa = n
	}
}

// BatchSize is an option for customizing the number of pixel to run as
// a concurrent task. By default the number of batch size is 32 (heuristic).
func BatchSize(n int32) Opt {
	return func(r *Renderer) {
		r.batchSize = n
	}
}

// Workers is an option for customizing the number of internal workers
// for a renderer. By default the number of workers equals to the number
// of CPUs.
func Workers(n int) Opt {
	return func(r *Renderer) {
		r.workers = n
	}
}

func Format(format buffer.PixelFormat) Opt {
	return func(r *Renderer) {
		r.format = format
	}
}

// Options updates the settings of the given renderer.
// If the renderer is running, the function waits until the rendering
// task is complete.
func (r *Renderer) Options(opts ...Opt) {
	r.wait() // wait last frame to finish

	for _, opt := range opts {
		opt(r)
	}

	r.bufs = make([]*buffer.Buffer, r.buflen)
	r.resetBufs()
}
