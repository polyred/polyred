// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import "image/color"

type BasicMaterial struct {
	color color.RGBA
}

type BasicMaterialOption func(m *BasicMaterial)

func WithColor(c color.RGBA) BasicMaterialOption {
	return func(m *BasicMaterial) {
		m.color = c
	}
}
