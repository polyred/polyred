// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/shader"
)

var (
	Draw = func(r *Renderer, mvp *shader.MVP, tri *primitive.Triangle) {
		r.draw(mvp, tri)
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
	DrawPrimitive = func(r *Renderer, buf *buffer.FragmentBuffer, tri *primitive.Triangle, p ...shader.Vertex) {
		r.drawPrimitive(buf, tri, p...)
	}
)
