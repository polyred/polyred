// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
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
// current object, one must compute o.ModelMatrix().Mul(modelMatrix) as the final matrix.
func IterObjects[S Iterator, T any](s S, iter func(o T, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if oo, ok := o.(T); ok {
			return iter(oo, modelMatrix)
		}
		return true
	})
}
