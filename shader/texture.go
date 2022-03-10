// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

var _ Program = &TextureShader{}

type TextureShader struct {
	ModelMatrix math.Mat4[float32]
	ViewMatrix  math.Mat4[float32]
	ProjMatrix  math.Mat4[float32]
	Texture     *buffer.Texture
}

func (s *TextureShader) Vertex(v *primitive.Vertex) *primitive.Vertex {
	v.Pos = s.ProjMatrix.MulM(s.ViewMatrix).MulM(s.ModelMatrix).MulV(v.Pos)
	return v
}

func (s *TextureShader) Fragment(frag *primitive.Fragment) color.RGBA {
	col := s.Texture.Query(0, frag.U, 1-frag.V)
	return col
}
