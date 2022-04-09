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

// Scene is a scene graph.
type Scene struct {
	// root group is a single group for
	// accessing all added group objects.
	root *Group
}

// NewScene creates a scene graph for the given objects.
// Each of given objects are considered as an individual
// transformation group that attached to the root group
func NewScene(objects ...object.Object[float32]) *Scene {
	s := &Scene{}

	rootGroup := newGroup("root")
	rootGroup.object = nil
	rootGroup.root = s
	s.root = rootGroup
	for _, obj := range objects {
		s.root.Add(obj)
	}
	return s
}

// Add adds given objects as a single group to the root scene group,
// then returns the root group.
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
		IterGroupObjects(s.root.children[i], func(o mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool {
			return iter(o, s.root.ModelMatrix().MulM(modelMatrix))
		})
	}
}

func (s *Scene) IterLights(iter func(l light.Light, modelMatrix math.Mat4[float32]) bool) {
	for i := range s.root.children {
		IterGroupObjects(s.root.children[i], func(o light.Light, modelMatrix math.Mat4[float32]) bool {
			return iter(o, s.root.ModelMatrix().MulM(modelMatrix))
		})
	}
}

func (s *Scene) Lights() ([]light.Source, []light.Environment) {
	sources, envs := []light.Source{}, []light.Environment{}
	s.IterLights(func(ll light.Light, modelMatrix math.Mat4[float32]) bool {
		switch l := ll.(type) {
		case light.Source:
			sources = append(sources, l)
		case light.Environment:
			envs = append(envs, l)
		}
		return true
	})
	return sources, envs
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

func (s *Scene) Normalize() {
	aabb := &primitive.AABB{
		Min: math.NewVec3[float32](0, 0, 0),
		Max: math.NewVec3[float32](0, 0, 0),
	}
	s.IterObjects(func(m object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		aabb.Add(m.AABB())
		return true
	})
	center := aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
	radius := aabb.Max.Sub(aabb.Min).Len() / 2
	s.IterObjects(func(m object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		m.Translate(-center.X, -center.Y, -center.Z)
		m.Scale(1/radius, 1/radius, 1/radius)
		return true
	})
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

func newGroup(name string) *Group {
	g := &Group{
		name:     name,
		root:     nil,
		object:   nil,
		parent:   nil,
		children: []*Group{},
	}
	g.ResetContext()
	return g
}

func (g *Group) Name() string      { return g.name }
func (g *Group) Type() object.Type { return object.TypeGroup }

// Add adds given objects as a new group to the given group.
func (g *Group) Add(geo ...object.Object[float32]) *Group {
	for i := range geo {
		gg := newGroup(geo[i].Name())
		gg.root = g.root
		gg.parent = g
		gg.object = geo[i]
		g.children = append(g.children, gg)
	}
	return g
}

func (g *Group) IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	if g.object != nil {
		iter(g.object, g.ModelMatrix().MulM(g.object.ModelMatrix()))
	}
	for i := range g.children {
		g.children[i].IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
			return iter(o, g.ModelMatrix().MulM(o.ModelMatrix()))
		})
	}
}

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

func IterObjects[T any](s *Scene, iter func(T, math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		switch any(o).(type) {
		case T:
			return iter(o.(T), modelMatrix)
		default:
			return true
		}
	})
}

func IterGroupObjects[T any](g *Group, iter func(T, math.Mat4[float32]) bool) {
	g.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		switch any(o).(type) {
		case T:
			return iter(o.(T), modelMatrix)
		default:
			return true
		}
	})
}
