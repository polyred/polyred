// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"poly.red/buffer"
	"poly.red/color"
	"poly.red/internal/alloc"
)

type Material struct {
	ID               uint64
	FlatShading      bool
	AmbientOcclusion bool
	ReceiveShadow    bool
	Texture          *buffer.Texture
}

type BlinnPhong struct {
	Material
	Ambient   color.RGBA
	Diffuse   color.RGBA
	Specular  color.RGBA
	Emissive  color.RGBA
	Shininess float32
	Opacity   float32
}

func NewBlinnPhong(opts ...BlinnPhongOption) *BlinnPhong {
	t := &BlinnPhong{
		Material: Material{
			ID:               alloc.ID(),
			FlatShading:      false,
			ReceiveShadow:    false,
			AmbientOcclusion: false,
			Texture:          nil,
		},
		Diffuse:   color.FromValue(0.5, 0.5, 0.5, 1.0),
		Specular:  color.FromValue(0.5, 0.5, 0.5, 1.0),
		Shininess: 1,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}
