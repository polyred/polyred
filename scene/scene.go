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

func (s *Scene) IterMeshes(iter func(m mesh.Mesh[float32], modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if o.Type() != object.TypeMesh {
			return true
		}

		return iter(o.(mesh.Mesh[float32]), s.root.ModelMatrix().MulM(modelMatrix))
	})
}

func (s *Scene) IterLights(iter func(l light.Light, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if o.Type() != object.TypeLight {
			return true
		}

		return iter(o.(light.Light), s.root.ModelMatrix().MulM(modelMatrix))
	})
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
