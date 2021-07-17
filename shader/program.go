// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader

import (
	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

type (
	// VertexProgram is a shader that executes on each vertex.
	VertexProgram func(primitive.Vertex) primitive.Vertex

	// FragmentShader is a shader that executes on each pixel.
	FragmentProgram func(primitive.Fragment) color.RGBA
)

// Program is a interface that describes a pair of shader programs.
type Program interface {
	VertexShader(primitive.Vertex) primitive.Vertex
	FragmentShader(primitive.Fragment) color.RGBA
}

var _ Program = &BasicShader{}

// BasicShader is a shader that does the minimum shading.
type BasicShader struct {
	ModelMatrix      math.Mat4
	ViewMatrix       math.Mat4
	ProjectionMatrix math.Mat4
}

func (s *BasicShader) VertexShader(v primitive.Vertex) primitive.Vertex {
	v.Pos = s.ProjectionMatrix.MulM(s.ViewMatrix).MulM(s.ModelMatrix).MulV(v.Pos)
	return v
}

func (s *BasicShader) FragmentShader(frag primitive.Fragment) color.RGBA {
	return frag.Col
}
