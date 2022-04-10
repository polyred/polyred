// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"poly.red/buffer"
	"poly.red/color"
	"poly.red/internal/alloc"
)

type Material interface {
	ID() uint64
}

type Standard struct {
	id               uint64
	FlatShading      bool
	AmbientOcclusion bool
	ReceiveShadow    bool
	Texture          *buffer.Texture
}

func (m *Standard) ID() uint64 { return m.id }

type BlinnPhong struct {
	Standard
	Ambient   color.RGBA
	Diffuse   color.RGBA
	Specular  color.RGBA
	Emissive  color.RGBA
	Shininess float32
	Opacity   float32
}

func NewBlinnPhong(opts ...BlinnPhongOption) *BlinnPhong {
	t := &BlinnPhong{
		Standard: Standard{
			id:               alloc.ID(),
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

func (m *BlinnPhong) ID() uint64 { return m.Standard.ID() }
