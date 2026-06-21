// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"poly.red/buffer"
	"poly.red/color"
)

// Default is the fallback material: a Blinn-Phong with a blue texture, used when
// geometry has no material of its own (for example OBJ faces with no .mtl
// entry). It is a single shared instance; do not mutate it.
func Default() *BlinnPhong { return defaultMaterial }

var defaultMaterial = &BlinnPhong{
	Standard: Standard{
		FlatShading:      false,
		AmbientOcclusion: false,
		ReceiveShadow:    false,
		Texture:          buffer.NewUniformTexture(color.Blue),
		name:             "default",
	},
	Ambient:   color.FromValue(0.7, 0.7, 0.7, 1.0),
	Diffuse:   color.FromValue(0.7, 0.7, 0.7, 1.0),
	Specular:  color.FromValue(0.5, 0.5, 0.5, 1.0),
	Shininess: 30.0,
}
