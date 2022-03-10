// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/material"
	"poly.red/shader"
)

var (
	Draw = func(r *Renderer, mvp *shader.MVP, tri *primitive.Triangle, m material.Material) {
		r.draw(mvp, tri, m)
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
	DrawPrimitive = func(r *Renderer, buf *buffer.FragmentBuffer, t1, t2, t3 *primitive.Vertex, p ...shader.Vertex) {
		r.drawPrimitive(buf, t1, t2, t3, p...)
	}
)
