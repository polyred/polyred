// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/camera"
	"poly.red/scene"
)

type option struct {
	BatchSize     int
	Workers       int
	Width, Height int
	Format        buffer.PixelFormat
	Background    color.RGBA
	MSAA          int
	Perspect      bool
	ShadowMap     bool
	GammaCorrect  bool
	Debug         bool
	Camera        camera.Interface
	Scene         *scene.Scene
	BlendFunc     BlendFunc
}

// Option represents a rendering option
type Option func(r *option)

func Size(width, height int) Option {
	return func(o *option) {
		o.Width = width
		o.Height = height
	}
}

func PixelFormat(format buffer.PixelFormat) Option {
	return func(o *option) { o.Format = format }
}

func Camera(cam camera.Interface) Option {
	return func(o *option) {
		o.Camera = cam
		if _, ok := cam.(*camera.Perspective); ok {
			o.Perspect = true
		}
	}
}

func Scene(s *scene.Scene) Option {
	return func(o *option) { o.Scene = s }
}

func Background(c color.RGBA) Option {
	return func(o *option) { o.Background = c }
}

func MSAA(n int) Option {
	return func(o *option) { o.MSAA = n }
}

// ShadowMap is an option that customizes whether to use shadow map or not.
func ShadowMap(enable bool) Option {
	return func(o *option) { o.ShadowMap = enable }
}

// GammaCorrection is an option that customizes whether gamma correction
// should be applied or not.
func GammaCorrection(enable bool) Option {
	return func(o *option) { o.GammaCorrect = enable }
}

// Blending is an option that customizes the blend function.
func Blending(f BlendFunc) Option {
	return func(o *option) { o.BlendFunc = f }
}

// Debug is an option that activates the debugging information of
// the given renderer.
//
// TODO: allow set a logger.
func Debug(enable bool) Option {
	return func(o *option) { o.Debug = enable }
}

// BatchSize is an option for customizing the number of pixel to run as
// a concurrent task. By default the number of batch size is 32 (heuristic).
func BatchSize(n int) Option {
	return func(o *option) { o.BatchSize = n }
}

// Workers is an option for customizing the number of internal workers
// for a renderer. By default the number of workers equals to the number
// of CPUs.
func Workers(n int) Option {
	return func(o *option) { o.Workers = n }
}

// Options updates the settings of the given renderer.
// If the renderer is running, the function waits until the rendering
// task is complete.
func (r *Renderer) Options(opts ...Option) {
	r.wait() // wait last frame to finish

	for _, opt := range opts {
		opt(r.cfg)
	}

	r.bufs = make([]*buffer.FragmentBuffer, r.buflen)
	r.resetBufs()

	// initialize shadow maps
	if r.cfg.Scene != nil && r.cfg.ShadowMap {
		r.initShadowMaps()
		r.bufs[0].ClearFragment()
		r.bufs[0].ClearColor()
	}
}
