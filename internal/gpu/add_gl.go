// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin

package gpu

import (
	"poly.red/internal/driver/gl"
	"poly.red/math"
)

var device gl.Device

func init() {

}

// add is a GPU version of math.Mat[float32].Add method.
func add[T math.Type](m1, m2 math.Mat[T]) math.Mat[T] {
	panic("unimplemented")
}

type shaderFn struct{}
