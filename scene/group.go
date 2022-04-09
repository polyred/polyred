// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene/object"
)

var _ object.Object[float32] = &Group{}

func (g *Group) Name() string { return g.name }

func (g *Group) SetName(name string) { g.name = name }

func (g *Group) Type() object.Type { return object.TypeGroup }

func (g *Group) AABB() primitive.AABB {
	aabb := &primitive.AABB{
		Min: math.NewVec3[float32](0, 0, 0),
		Max: math.NewVec3[float32](0, 0, 0),
	}
	g.IterObjects(func(m object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		aabb.Add(m.AABB())
		return true
	})
	return *aabb
}

func (g *Group) Normalize() {
	aabb := g.AABB()
	center := aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	g.IterObjects(func(m object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		m.Translate(-center.X, -center.Y, -center.Z)
		m.Scale(1/radius, 1/radius, 1/radius)
		return true
	})
}
