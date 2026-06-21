// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shader

import (
	"poly.red/color"
	"poly.red/geometry/primitive"
)

// Fragment is a per-fragment shading function. The renderer applies it over a
// fragment buffer via Renderer.DrawFragments (used for deferred shading and
// gamma correction).
type Fragment func(*primitive.Fragment) color.RGBA
