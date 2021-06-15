// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
)

var (
	Draw = func(r *Renderer, uniforms map[string]math.Matrix, tri *geometry.Triangle, mat material.Material) {
		r.draw(uniforms, tri, mat)
	}
	ResetBuf = func(r *Renderer) {
		r.resetBufs()
	}
	ForwardPass = func(r *Renderer) {
		r.forwardPass()
	}
	DeferredPass = func(r *Renderer) {
		r.deferredPass()
	}
)
