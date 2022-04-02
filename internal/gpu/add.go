// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gpu

import (
	_ "embed"

	"poly.red/math"
)

// Add is a GPU version of math.Mat[float32].Add method.
func Add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	return add(m1, m2)
}
