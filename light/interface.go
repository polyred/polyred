// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/math"
	"poly.red/object"
)

// Environment represents the abstraction of environment lighting.
// Such as ambient light, etc.
type Environment interface {
	object.Object

	Color() color.RGBA
	Intensity() float32
}

// Source represents the abstraction of a light source.
type Source interface {
	object.Object

	Color() color.RGBA
	Intensity() float32
	Position() math.Vec3
	CastShadow() bool
}
