// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package rend

import (
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/material"
	"changkun.de/x/ddd/math"
)

var (
	Draw = func(r *Rasterizer, uniforms map[string]math.Matrix, tri *geometry.Triangle, mat material.Material) {
		r.draw(uniforms, tri, mat)
	}
	ResetBuf = func(r *Rasterizer) {
		r.resetBufs()
	}
	ForwardPass = func(r *Rasterizer) {
		r.forwardPass()
	}
	DeferredPass = func(r *Rasterizer) {
		r.deferredPass()
	}
)
