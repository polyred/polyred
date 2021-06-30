// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"changkun.de/x/polyred/camera"
	"changkun.de/x/polyred/geometry"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

// Scene is a scene graph
type Scene struct {
	root   *Group
	camera camera.Interface
}

func NewScene() *Scene {
	s := &Scene{}

	rootGroup := newGroup()
	rootGroup.name = "root"
	rootGroup.object = nil
	rootGroup.root = s
	s.root = rootGroup

	return s
}

func (s *Scene) Add(geo ...object.Object) *Group {
	s.root.Add(geo...)
	return s.root
}

func (s *Scene) SetCamera(c camera.Interface) {
	s.camera = c
}

func (s *Scene) GetCamera() camera.Interface {
	return s.camera
}

func (s *Scene) IterObjects(iter func(o object.Object, modelMatrix math.Mat4) bool) {

	for i := range s.root.children {
		s.root.children[i].IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
			iter(o, s.root.ModelMatrix().MulM(modelMatrix))
			return true
		})
	}
}

func (s *Scene) Center() math.Vec4 {
	aabb := &primitive.AABB{
		Min: math.NewVec4(0, 0, 0, 1),
		Max: math.NewVec4(0, 0, 0, 1),
	}
	s.IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
		if o.Type() != object.TypeMesh {
			return true
		}
		mesh := o.(geometry.Mesh)
		aabb.Add(mesh.AABB())
		return true
	})
	return aabb.Min.Add(aabb.Max).Pos()
}

var _ object.Object = &Group{}

// Group is a group of geometry objects, and also implements
// the geometry.Object interface.
type Group struct {
	math.TransformContext

	name     string
	root     *Scene
	object   object.Object
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

func (g *Group) Add(geo ...object.Object) *Group {
	for i := range geo {
		gg := newGroup()
		gg.root = g.root
		gg.parent = g
		gg.object = geo[i]
		g.children = append(g.children, gg)
	}

	return g
}

func (g *Group) IterObjects(iter func(o object.Object, modelMatrix math.Mat4) bool) {
	iter(g.object, g.ModelMatrix().MulM(g.object.ModelMatrix()))
	for i := range g.children {
		g.children[i].IterObjects(func(o object.Object, modelMatrix math.Mat4) bool {
			return iter(o, g.ModelMatrix().MulM(o.ModelMatrix()))
		})
	}
}
