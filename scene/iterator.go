// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/math"
	"poly.red/scene/object"
)

// Iterator represents an iterable scene object.
type Iterator interface {
	IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool)
}

// IterObjects traverses all objects over a given scene graph or a scene group or any type
// that implements Iterator interface. The user defined iter function receives the model
// matrix of the previous group transformation. To obtain the correct model matrix of the
// current object, one must compute modelMatrix.MulM(o.ModelMatrix()) as the final matrix.
func IterObjects[S Iterator, T any](s S, iter func(o T, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if oo, ok := o.(T); ok {
			return iter(oo, modelMatrix)
		}
		return true
	})
}

// IterVisibleGeometry traverses only visible objects that are inside the view frustum of
// the given camera.
func IterVisibleGeometry[S Iterator](s S, c camera.Interface, iter func(g *geometry.Geometry, modelMatrix math.Mat4[float32]) bool) {
	panic("umimplemented")
}
