// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"
	"math/rand"

	"poly.red/internal/deepcopy"
	"poly.red/math"
)

// AttrName is a string that represents the name of the attribute.
type AttrName string

var (
	AttrPosition AttrName = "pos"
	AttrNormal   AttrName = "nor"
	AttrColor    AttrName = "col"
	AttrUV       AttrName = "uv"
	AttrMVP      AttrName = "mvp"
)

// Vertex represents a vertex that conveys attributes.
type Vertex struct {
	Idx int
	Pos math.Vec4[float32] // Position is the vertex position
	Nor math.Vec4[float32] // Nor is the vertex normal
	Col color.RGBA         // Col is the vertex color
	UV  math.Vec2[float32] // UV is the vertex UV coordinates

	// AttrSmooth is interpolated between vertex and fragment shaders.
	AttrSmooth map[AttrName]any
	// AttrFlat is not interpolated between vertex and fragment shaders.
	AttrFlat map[AttrName]any
}

// NewVertex creates a new Vertex and have an unset index.
func NewVertex(opts ...VertOpt) *Vertex {
	v := &Vertex{
		AttrFlat:   map[AttrName]any{},
		AttrSmooth: map[AttrName]any{},
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

type VertOpt func(v *Vertex)

func Idx(i int) VertOpt {
	return func(v *Vertex) {
		v.Idx = i
	}
}
func Pos(pos math.Vec4[float32]) VertOpt {
	return func(v *Vertex) {
		v.Pos = pos
	}
}
func UV(uv math.Vec2[float32]) VertOpt {
	return func(v *Vertex) {
		v.UV = uv
	}
}
func Nor(nor math.Vec4[float32]) VertOpt {
	return func(v *Vertex) {
		v.Nor = nor
	}
}
func Col(col color.RGBA) VertOpt {
	return func(v *Vertex) {
		v.Col = col
	}
}

// NewRandomVertex returns a vertex that its position, normal, color and
// UV coordinates are randomly generated.
func NewRandomVertex() *Vertex {
	return NewVertex(
		Pos(math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 1)),
		Nor(math.NewVec4(rand.Float32(), rand.Float32(), rand.Float32(), 0).Unit()),
		Col(color.RGBA{uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int()), uint8(rand.Int())}),
		UV(math.NewVec2(rand.Float32(), rand.Float32())),
	)
}

// Copy returns a deep copy of the current vertex.
func (v *Vertex) Copy() *Vertex {
	u := NewVertex(
		Idx(v.Idx),
		Pos(v.Pos),
		Nor(v.Nor),
		Col(v.Col),
		UV(v.UV),
	)

	for k, attr := range v.AttrSmooth {
		u.AttrSmooth[k] = deepcopy.Value(attr)
	}
	for k, attr := range v.AttrFlat {
		u.AttrFlat[k] = deepcopy.Value(attr)
	}
	return u
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos.ToVec3(),
		Max: v.Pos.ToVec3(),
	}
}
