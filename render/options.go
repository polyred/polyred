// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"
	"image/color"
	"sync"

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

func ShadowMap(enable bool) Opt {
	return func(r *Renderer) {
		r.useShadowMap = enable
	}
}

func GammaCorrection(enable bool) Opt {
	return func(r *Renderer) {
		r.correctGamma = enable
	}
}

func Blending(f BlendFunc) Opt {
	return func(r *Renderer) {
		r.blendFunc = f
	}
}

func Debug(enable bool) Opt {
	return func(r *Renderer) {
		r.debug = enable
	}
}

func Concurrency(n int32) Opt {
	return func(r *Renderer) {
		r.concurrentSize = n
	}
}

func ThreadLimit(n int) Opt {
	return func(r *Renderer) {
		r.gomaxprocs = n
	}
}

func (r *Renderer) Options(opts ...Opt) {
	r.wait() // wait last frame to finish

	for _, opt := range opts {
		opt(r)
	}

	w := r.width * r.msaa
	h := r.height * r.msaa

	// calibrate rendering size
	r.lockBuf = make([]sync.Mutex, w*h)
	r.gBuf = make([]gInfo, w*h)
	r.frameBuf = image.NewRGBA(image.Rect(0, 0, w, h))

	r.lightSources = []light.Source{}
	r.lightEnv = []light.Environment{}
	if r.scene != nil {
		r.scene.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
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
	}

	r.resetGBuf()
	r.resetFrameBuf()
}
