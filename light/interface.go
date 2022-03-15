// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"poly.red/math"
	"poly.red/scene/object"
)

// Light represents a colored light.
type Light interface {
	object.Object[float32]
	Color() color.RGBA
}

// Environment represents the abstraction of environment lighting.
// Such as ambient light, etc.
type Environment interface {
	object.Object[float32]

	Color() color.RGBA
	Intensity() float32
}

// Source represents the abstraction of a light source.
type Source interface {
	object.Object[float32]

	Color() color.RGBA
	Intensity() float32
	Position() math.Vec3[float32]
	CastShadow() bool
}
