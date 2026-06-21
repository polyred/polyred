// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"

	"poly.red/math"
)

// Fragment represents a shaded pixel and its interpolated attributes.
type Fragment struct {
	X, Y       int
	Depth      float32
	U, V       float32
	Du, Dv     float32
	Nor        math.Vec4[float32]
	Col        color.RGBA
	MaterialID int64
	FaceNor    math.Vec4[float32]
	WordPos    math.Vec4[float32]
}
