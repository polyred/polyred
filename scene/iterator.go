// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package scene

import (
	"poly.red/math"
	"poly.red/scene/object"
)

type Iterator interface {
	IterObjects(iter func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool)
}

// Iterator traverses objects over a scene or a scene group or any type
// that implements Iterator interface.
func IterObjects[O Iterator, T any](s O, iter func(o T, modelMatrix math.Mat4[float32]) bool) {
	s.IterObjects(func(o object.Object[float32], modelMatrix math.Mat4[float32]) bool {
		if oo, ok := o.(T); ok {
			return iter(oo, modelMatrix)
		}
		return true
	})
}
