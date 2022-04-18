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
	var (
		aabb *primitive.AABB
		n    int
	)
	g.IterObjects(func(m object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if n == 0 {
			initial := m.AABB()
			aabb = &initial
		} else {
			aabb.Add(m.AABB())
		}
		n++
		return true
	})

	// If the current group does not contain any objects, return the origin's AABB.
	if n == 0 {
		return primitive.AABB{}
	}

	return *aabb
}

// Normalize normalizes objects in this scene group.
func (g *Group) Normalize() {
	modelMatrix := g.ModelMatrix()
	aabb := g.AABB()

	// A group of objects may already being transformed by its
	// group model matrix. To correctly compute the correct size
	// of a group of objects, compute its AABB then learn the actual
	// size of this group of object, then compute the center and radius.
	min := modelMatrix.MulV(aabb.Min.ToVec4(1)).ToVec3()
	max := modelMatrix.MulV(aabb.Max.ToVec4(1)).ToVec3()
	center := min.Add(max).Scale(0.5, 0.5, 0.5)
	radius := max.Sub(min).Len() / 2
	fac := 1 / radius
	g.Translate(-center.X, -center.Y, -center.Z)
	g.Scale(fac, fac, fac)
}
