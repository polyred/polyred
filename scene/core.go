// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package scene manages a scene graph.
//
// Any object.Object[float32] can be added to the scene graph.
package scene

import (
	"errors"

	"poly.red/math"
	"poly.red/scene/object"
)

var errStop = errors.New("iteration stop")

// Scene represents a scene graph.
//
// A scene graph is a hierarchical tree where the model matrix
// of lower-level objects is equal to its model matrix multiplied
// by higher level parent group model matrix on the left side.
//
//             Scene
//            /  |   \
//           /   |    \
//      Group  Object  Object
//        / \
//       /   \
//    Object  Object
type Scene struct {
	// root group is a single group for accessing
	// all added objects.
	//
	// The root group holds a global transformation
	// matrix of all added objects.
	root *Group
}

// NewScene creates a scene graph for the given objects.
// Each of given objects are considered as an individual
// transformation group that attached to the root group.
func NewScene(objects ...object.Object[float32]) *Scene {
	s := &Scene{}
	s.root = NewGroup()
	s.root.name = "scene_root"
	s.root.root = s
	s.root.objects = nil // always nil in a scene's root group.
	for i := range objects {
		// If the adding object is a group, then assign the
		// current root and the current parent.
		if g, ok := objects[i].(*Group); ok {
			g.root = s
			g.parent = s.root
		}

		// Record the adding object. This will record the object in allObjects.
		s.root.Add(objects[i])
	}
	return s
}

// Add adds given objects as a single group to the root scene group,
// then returns the root group.
//
// If the same object being added to the group multiple times, it will
// still be considered as multiple objects and will be iterated multiple
// times.
func (s *Scene) Add(objects ...object.Object[float32]) *Group {
	for i := range objects {
		// If the adding object is a group object, we need properly
		// assign its root scene graph (s) and its parents (s.root).
		if g, ok := objects[i].(*Group); ok {
			g.root = s
			g.parent = s.root
		}

		// Record the adding object.
		s.root.Add(objects[i])
	}
	return s.root
}

// IterObjects traverse all objects in this scene graph.
func (s *Scene) IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, errStop) {
				return
			}
			panic(r)
		}
	}()

	for i := range s.root.objects {
		if gg, ok := s.root.objects[i].(*Group); ok {
			gg.iterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
				if ok := iter(o, s.root.ModelMatrix().MulM(modelMatrix)); ok {
					return true
				}
				panic(errStop)
			})
		} else {
			if ok := iter(s.root.objects[i], s.root.ModelMatrix()); ok {
				continue
			}
			break
		}
	}
}

// Group is a group of geometry objects, and also implements
// the geometry.Object interface.
type Group struct {
	math.TransformContext[float32]

	name    string
	root    *Scene
	parent  *Group
	objects []object.Object[float32]
}

// NewGroup creates a new groupped objects.
func NewGroup() *Group {
	g := &Group{
		name:    "group",
		root:    nil,
		parent:  nil,
		objects: []object.Object[float32]{},
	}
	g.ResetContext()
	return g
}

// Add adds a list of given objects. The added objects are belongs to this
// group, therefore if the model matrix of the given group is M1, and an object
// in this group is another group. Then the model matrix of models in the other
// group is computed as M1*M2. The returned group is the given group.
//
// If the same object being added to the group multiple times, it will still be
// considered as multiple objects and will be iterated multiple times.
func (g *Group) Add(objects ...object.Object[float32]) *Group {
	for i := range objects {
		if gg, ok := objects[i].(*Group); ok {
			gg.parent = g
			gg.root = g.root
		}

		// Finally, add the object to the group objects.
		g.objects = append(g.objects, objects[i])
	}
	return g
}

// IterObjects traverse all objects in this group.
func (g *Group) IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok && errors.Is(err, errStop) {
				return
			}
			panic(r)
		}
	}()
	for i := range g.objects {
		if gg, ok := g.objects[i].(*Group); ok {
			gg.iterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
				if ok := iter(o, g.ModelMatrix().MulM(modelMatrix)); ok {
					return true
				}
				panic(errStop)
			})
		} else {
			if ok := iter(g.objects[i], g.ModelMatrix().MulM(g.objects[i].ModelMatrix())); ok {
				continue
			}
			panic(errStop)
		}
	}
}

// iterObjects is the underlying recursive iterator for the given group.
// If iter returns false. This method will panic with an errStop. This error
// will be captured by the higher level IterObjects so that we can properly
// stop the entire traversal at once.
func (g *Group) iterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool) {
	for i := range g.objects {
		if gg, ok := g.objects[i].(*Group); ok {
			gg.iterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
				if ok := iter(o, g.ModelMatrix().MulM(modelMatrix)); ok {
					return true
				}
				panic(errStop)
			})
		} else {
			if ok := iter(g.objects[i], g.ModelMatrix().MulM(g.objects[i].ModelMatrix())); ok {
				continue
			}
			panic(errStop)
		}
	}
}
