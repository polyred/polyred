// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"fmt"
	"image/color"
	"math/rand"

	"poly.red/internal/deepcopy"
	"poly.red/math"
)

// Attribute is a string that represents the name of the attribute.
type Attribute string

// Vertex represents a vertex that conveys attributes.
type Vertex struct {
	Idx uint64
	Pos math.Vec4  // Position is the vertex position
	Nor math.Vec4  // Nor is the vertex normal
	Col color.RGBA // Col is the vertex color
	UV  math.Vec2  // UV is the vertex UV coordinates

	// AttrSmooth is interpolated between vertex and fragment shaders.
	AttrSmooth map[Attribute]any
	// AttrFlat is not interpolated between vertex and fragment shaders.
	AttrFlat map[Attribute]any
}

// NewVertex creates a new Vertex and have an unset index.
func NewVertex(pos math.Vec4) *Vertex {
	return &Vertex{Pos: pos}
}

// NewRandomVertex returns a vertex that its position, normal, color and
// UV coordinates are randomly generated.
func NewRandomVertex() *Vertex {
	return &Vertex{
		Pos: math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 1),
		Nor: math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 0).Unit(),
		Col: color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())},
		UV:  math.NewVec2(rand.Float32(), rand.Float32()),
	}
}

// Copy returns a deep copy of the current vertex.
func (v *Vertex) Copy() *Vertex {
	u := &Vertex{
		Idx: v.Idx,
		Pos: v.Pos,
		Nor: v.Nor,
		Col: v.Col,
		UV:  v.UV,
	}

	var err error
	for k, attr := range v.AttrSmooth {
		u.AttrSmooth[k], err = deepcopy.Anything(attr)
		if err != nil {
			panic(fmt.Sprintf("primitive: vertex %d stored non-copiable content in AttrSmooth: %v", v.Idx, k))
		}
	}
	for k, attr := range v.AttrFlat {
		u.AttrFlat[k], err = deepcopy.Anything(attr)
		if err != nil {
			panic(fmt.Sprintf("primitive: vertex %d stored non-copiable content in AttrFlat: %v", v.Idx, k))
		}
	}
	return u
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos.ToVec3(),
		Max: v.Pos.ToVec3(),
	}
}
