// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image"
	"image/color"
)

func NewBasicMaterial(c color.Color) Material {
	data := image.NewRGBA(image.Rect(0, 0, 1, 1))
	data.Set(0, 0, c)
	tex := NewTexture(
		WithImage(data),
		WithIsotropicMipMap(true),
	)
	return NewBlinnPhong(
		WithBlinnPhongTexture(tex),
		WithBlinnPhongFactors(0.6, 1),
		WithBlinnPhongShininess(100),
	)
}
