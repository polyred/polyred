// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/light"
	"poly.red/math"
	"poly.red/object"
	"poly.red/scene"
)

// Opt represents a rendering Opt
type Opt func(r *Renderer)

func Size(width, height int) Opt {
	return func(r *Renderer) {
		r.width = width
		r.height = height
	}
}

func PixelFormat(format buffer.PixelFormat) Opt {
	return func(r *Renderer) {
		r.pixelFormat = format
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

// ShadowMap is an option that customizes whether to use shadow map or not.
func ShadowMap(enable bool) Opt {
	return func(r *Renderer) {
		r.useShadowMap = enable
	}
}

// GammaCorrection is an option that customizes whether gamma correction
// should be applied or not.
func GammaCorrection(enable bool) Opt {
	return func(r *Renderer) {
		r.correctGamma = enable
	}
}

// Blending is an option that customizes the blend function.
func Blending(f BlendFunc) Opt {
	return func(r *Renderer) {
		r.blendFunc = f
	}
}

// Debug is an option that activates the debugging information of
// the given renderer.
//
// TODO: allow set a logger.
func Debug(enable bool) Opt {
	return func(r *Renderer) {
		r.debug = enable
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

// Options updates the settings of the given renderer.
// If the renderer is running, the function waits until the rendering
// task is complete.
func (r *Renderer) Options(opts ...Opt) {
	r.wait() // wait last frame to finish

	for _, opt := range opts {
		opt(r)
	}

	r.bufs = make([]*buffer.FragmentBuffer, r.buflen)
	r.resetBufs()
	r.lightSources = []light.Source{}
	r.lightEnv = []light.Environment{}
	if r.scene != nil {
		r.scene.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			if o.Type() != object.TypeLight {
				return true
			}

			switch l := o.(type) {
			case light.Source:
				r.lightSources = append(r.lightSources, l)
			case light.Environment:
				r.lightEnv = append(r.lightEnv, l)
			}
			return true
		})
	}

	// initialize shadow maps
	if r.scene != nil && r.useShadowMap {
		r.initShadowMaps()
		r.bufs[0].ClearFragment()
		r.bufs[0].ClearColor()
	}
}
