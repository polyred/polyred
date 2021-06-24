// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import (
	"image/color"

	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

// Environment represents the abstraction of environment lighting.
// Such as ambient light, etc.
type Environment interface {
	object.Object

	Color() color.RGBA
	Intensity() float64
}

// Source represents the abstraction of a light source.
type Source interface {
	object.Object

	Color() color.RGBA
	Intensity() float64
	Position() math.Vector
	CastShadow() bool
}
