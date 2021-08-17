// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/math"
)

var (
	Draw = func(r *Renderer, uniforms map[string]interface{}, tri *primitive.Triangle, modelMatrix math.Mat4, m material.Material) {
		r.draw(uniforms, tri, m)
	}
	PassForward = func(r *Renderer) {
		r.passForward()
	}
	PassDeferred = func(r *Renderer) {
		r.passDeferred()
	}
	PassAntiAliasing = func(r *Renderer) {
		r.passAntialiasing()
	}
	PassGammaCorrect = func(r *Renderer) {
		r.correctGamma = true
		r.passGammaCorrect()
	}
)
