// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"
	"math/rand"

	"poly.red/math"
)

// Vertex represents a vertex that conveys attributes.
type Vertex struct {
	Pos        math.Vec4
	UV         math.Vec4
	Nor        math.Vec4
	Col        color.RGBA
	AttrSmooth map[string]any
	AttrFlat   map[string]any
}

// Fragment represents a pixel that conveys varied attributes.
type Fragment struct {
	X, Y       int
	Depth      float32
	UV         math.Vec2
	Du         float32
	Dv         float32
	Nor        math.Vec4
	Col        color.RGBA
	AttrSmooth map[string]any
	AttrFlat   map[string]any
}

func NewRandomVertex() *Vertex {
	return &Vertex{
		Pos: math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 1),
		UV:  math.NewVec4(rand.Float32(), rand.Float32(), 0, 1),
		Nor: math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 1).Unit(),
		Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())},
	}
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos.ToVec3(),
		Max: v.Pos.ToVec3(),
	}
}
