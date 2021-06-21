// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"
	"math/rand"

	"changkun.de/x/ddd/math"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Pos math.Vector
	UV  math.Vector
	Nor math.Vector
	Col color.RGBA
}

func NewRandomVertex() *Vertex {
	return &Vertex{
		Pos: math.NewVector(rand.Float64(), rand.Float64(), rand.Float64(), 1),
		UV:  math.NewVector(rand.Float64(), rand.Float64(), 0, 1),
		Nor: math.NewVector(rand.Float64(), rand.Float64(), rand.Float64(), 1).Unit(),
		Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())},
	}
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos,
		Max: v.Pos,
	}
}
