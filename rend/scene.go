// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
)

// Scene represents a basic scene graph
type Scene struct {
	Meshes []*geometry.TriangleMesh
	Lights []light.Light
	Camera camera.Interface

	aabb *primitive.AABB
}

// NewScene returns a new scene graph
func NewScene() *Scene {
	return &Scene{}
}

// AddMesh adds a mesh to the scene graph
func (s *Scene) AddMesh(m *geometry.TriangleMesh) {
	s.Meshes = append(s.Meshes, m)
}

// AddLight adds a light to the scene graph
func (s *Scene) AddLight(l light.Light) {
	s.Lights = append(s.Lights, l)
}

// UseCamera uses the given camera for rendering scene graph
func (s *Scene) UseCamera(c camera.Interface) {
	s.Camera = c
}

func (s *Scene) AABB() primitive.AABB {
	if s.aabb == nil {
		s.aabb = &primitive.AABB{
			Min: math.NewVector(0, 0, 0, 1),
			Max: math.NewVector(0, 0, 0, 1),
		}
		for i := range s.Meshes {
			s.aabb.Add(s.Meshes[i].AABB())
		}
	}
	return *s.aabb
}

// Center returns the center of the scene
func (s *Scene) Center() math.Vector {
	aabb := s.AABB()
	return aabb.Min.Add(aabb.Max).Pos()
}
