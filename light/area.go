// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"poly.red/color"
	"poly.red/geometry"
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene/object"
)

var (
	_ Light                  = &Area{}
	_ Source                 = &Area{}
	_ object.Object[float32] = &Area{}
)

type Area struct {
	math.TransformContext[float32]

	position   math.Vec3[float32]
	intensity  float32
	color      color.RGBA
	shape      *geometry.Geometry
	maxBounces int
	castShadow bool
}

func NewArea(opts ...Option) Source {
	// FIXME: construct a geometry manually here to avoid circular import.
	// Find a better way to do this.
	width, height := float32(1.0), float32(1.0)
	v1 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(-0.5*width, 0, -0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](0, 1)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{255, 0, 0, 255}),
	)
	v2 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(-0.5*width, 0, 0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](0, 0)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 255, 0, 255}),
	)
	v3 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(0.5*width, 0, 0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](1, 0)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 0, 255, 255}),
	)
	v4 := primitive.NewVertex(
		primitive.Pos(math.NewVec4(0.5*width, 0, -0.5*height, 1)),
		primitive.UV(math.NewVec2[float32](1, 1)),
		primitive.Nor(math.NewVec4[float32](0, 1, 0, 0)),
		primitive.Col(color.RGBA{0, 0, 0, 255}),
	)
	plane := mesh.NewTriangleMesh([]*primitive.Triangle{
		{V1: v1, V2: v2, V3: v3},
		{V1: v1, V2: v3, V3: v4},
	})

	a := &Area{
		intensity:  0.1,
		color:      color.White,
		shape:      geometry.New(plane),
		maxBounces: 1024,
		castShadow: false,
	}
	for _, opt := range opts {
		opt(a)
	}
	a.ResetContext()
	return a
}

func (a *Area) Name() string                 { return "area_light" }
func (a *Area) Type() object.Type            { return object.TypeLight }
func (a *Area) Color() color.RGBA            { return a.color }
func (a *Area) Intensity() float32           { return a.intensity }
func (a *Area) AABB() primitive.AABB         { return primitive.NewAABB(math.NewVec3[float32](0, 0, 0)) }
func (a *Area) CastShadow() bool             { return a.castShadow }
func (a *Area) Position() math.Vec3[float32] { return a.position }
