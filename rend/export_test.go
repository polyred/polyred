// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/material"
	"changkun.de/x/polyred/math"
)

var (
	Draw = func(r *Renderer, uniforms map[string]interface{}, tri *primitive.Triangle, modelMatrix math.Matrix, m material.Material) {
		r.draw(uniforms, tri, modelMatrix, m)
	}
	ResetGBuf = func(r *Renderer) {
		r.resetGBuf()
	}
	ResetFrameBuf = func(r *Renderer) {
		r.resetFrameBuf()
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
)
