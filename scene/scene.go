// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/light"
	"poly.red/math"
	"poly.red/scene/object"
)

func (s *Scene) IterGeometry(iter func(m *geometry.Geometry, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if o.Type() != object.TypeGeometry {
			return true
		}

		return iter(o.(*geometry.Geometry), s.root.ModelMatrix().MulM(modelMatrix))
	})
}

func (s *Scene) IterLight(iter func(l light.Light, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if o.Type() != object.TypeLight {
			return true
		}

		return iter(o.(light.Light), s.root.ModelMatrix().MulM(modelMatrix))
	})
}

func (s *Scene) Lights() ([]light.Source, []light.Environment) {
	sources, envs := []light.Source{}, []light.Environment{}
	s.IterLight(func(ll light.Light, modelMatrix math.Mat4[float32]) bool {
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
	aabb := s.root.AABB()
	return aabb.Min.Add(aabb.Max).Scale(0.5, 0.5, 0.5)
}

func (s *Scene) AABB() primitive.AABB { return s.root.AABB() }
func (s *Scene) Normalize()           { s.root.Normalize() }
