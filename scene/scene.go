// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"poly.red/geometry/mesh"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
	"poly.red/scene/object"
)

// Scene is a scene graph
type Scene struct {
	root *Group
}

func NewScene(objects ...object.Object[float32]) *Scene {
	s := &Scene{}

	rootGroup := newGroup()
	rootGroup.name = "root"
	rootGroup.object = nil
	rootGroup.root = s
	s.root = rootGroup
	for _, obj := range objects {
		s.root.Add(obj)
	}
	return s
}

func (s *Scene) Add(geo ...object.Object[float32]) *Group {
	s.root.Add(geo...)
	return s.root
}

func (s *Scene) IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	for i := range s.root.children {
		s.root.children[i].IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			return iter(o, s.root.ModelMatrix().MulM(modelMatrix))
		})
	}
}

func (s *Scene) IterMeshes(iter func(m mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool) {
	for i := range s.root.children {
		s.root.children[i].IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			if o.Type() != object.TypeMesh {
				return true
			}

			return iter(o.(mesh.Mesh[float32]), s.root.ModelMatrix().MulM(modelMatrix))
		})
	}
}

func (s *Scene) IterLights(iter func(l light.Light, modelMatrix math.Mat4[float32]) bool) {
	for i := range s.root.children {
		s.root.children[i].IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			if o.Type() != object.TypeLight {
				return true
			}

			return iter(o.(light.Light), s.root.ModelMatrix().MulM(modelMatrix))
		})
	}
}

func (s *Scene) Center() math.Vec3[float32] {
	aabb := &primitive.AABB{
		Min: math.NewVec3[float32](0, 0, 0),
		Max: math.NewVec3[float32](0, 0, 0),
	}
	s.IterMeshes(func(m mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool {
		aabb.Add(m.AABB())
		return true
	})
	return aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
}

var _ object.Object[float32] = &Group{}

// Group is a group of geometry objects, and also implements
// the geometry.Object interface.
type Group struct {
	math.TransformContext[float32]

	name     string
	root     *Scene
	object   object.Object[float32]
	parent   *Group
	children []*Group
}

func newGroup() *Group {
	g := &Group{
		name:     "",
		root:     nil,
		object:   nil,
		parent:   nil,
		children: []*Group{},
	}
	g.ResetContext()
	return g
}

func (g *Group) Type() object.Type {
	return object.TypeGroup
}

func (g *Group) Add(geo ...object.Object[float32]) *Group {
	for i := range geo {
		gg := newGroup()
		gg.root = g.root
		gg.parent = g
		gg.object = geo[i]
		g.children = append(g.children, gg)
	}

	return g
}

func (g *Group) IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	iter(g.object, g.ModelMatrix().MulM(g.object.ModelMatrix()))
	for i := range g.children {
		g.children[i].IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			return iter(o, g.ModelMatrix().MulM(o.ModelMatrix()))
		})
	}
}
