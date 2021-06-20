// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"fmt"

	"changkun.de/x/ddd/camera"
	"changkun.de/x/ddd/geometry"
	"changkun.de/x/ddd/geometry/primitive"
	"changkun.de/x/ddd/light"
	"changkun.de/x/ddd/math"
)

// Scene represents a basic scene graph
type Scene struct {
	Name         string
	Meshes       []*geometry.TriangleMesh
	LightSources []light.Source
	LightEnv     []light.Environment
	Camera       camera.Interface

	aabb *primitive.AABB
}

// NewScene returns a new scene graph
func NewScene() *Scene {
	return &Scene{Name: "default_scene"}
}

// AddMesh adds a mesh to the scene graph
func (s *Scene) AddMesh(m *geometry.TriangleMesh) {
	s.Meshes = append(s.Meshes, m)
}

// AddLight is a wrapper of AddLightSource and AddLightEnvironment.
func (s *Scene) AddLight(ls ...interface{}) error {
	for _, l := range ls {
		switch ll := l.(type) {
		case *light.Point:
			s.AddLightSource(ll)
		case *light.Ambient:
			s.AddLightEnvironment(ll)
		default:
			return fmt.Errorf("unsupported light type: %v", ll)
		}
	}
	return nil
}

// AddLight adds a source light to the scene graph
func (s *Scene) AddLightSource(l ...light.Source) {
	s.LightSources = append(s.LightSources, l...)
}

// AddLightEnvironment adds an environment light to the scene graph
func (s *Scene) AddLightEnvironment(l ...light.Environment) {
	s.LightEnv = append(s.LightEnv, l...)
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
