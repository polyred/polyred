// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import (
	"image/color"
)

// Scene represents a scene graph
type Scene struct {
	Objects []*TriangleMesh
	Lights  []*PointLight
}

// NewScene returns a new scene graph
func NewScene() *Scene {
	return &Scene{}
}

// AddMesh adds a mesh to the scene graph
func (s *Scene) AddMesh(m *TriangleMesh) {
	s.Objects = append(s.Objects, m)
}

// AddLight adds a light to the scene graph
func (s *Scene) AddLight(l *PointLight) {
	s.Lights = append(s.Lights, l)
}

// PointLight is a point light
type PointLight struct {
	Color    color.RGBA
	Position Vector
	Kamb     float64
	Kdiff    float64
	Kspec    float64
}

// NewPointLight returns a new point light
func NewPointLight(c color.RGBA, p Vector, Ka, Kd, Ks float64) *PointLight {
	return &PointLight{
		Color:    c,
		Position: p,
		Kamb:     Ka,
		Kdiff:    Kd,
		Kspec:    Ks,
	}
}
