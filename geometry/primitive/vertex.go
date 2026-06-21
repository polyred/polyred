// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"
	"math/rand"

	"poly.red/math"
)

// Vertex represents a vertex with its rendering attributes.
type Vertex struct {
	Idx int
	Pos math.Vec4[float32] // Position is the vertex position
	Nor math.Vec4[float32] // Nor is the vertex normal
	Col color.RGBA         // Col is the vertex color
	UV  math.Vec2[float32] // UV is the vertex UV coordinates
}

// NewVertex creates a new Vertex and have an unset index.
func NewVertex(opts ...VertOpt) *Vertex {
	v := &Vertex{}
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
	return NewVertex(
		Idx(v.Idx),
		Pos(v.Pos),
		Nor(v.Nor),
		Col(v.Col),
		UV(v.UV),
	)
}

func (v *Vertex) AABB() AABB {
	return AABB{
		Min: v.Pos.ToVec3(),
		Max: v.Pos.ToVec3(),
	}
}
