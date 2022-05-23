// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"poly.red/buffer"
	"poly.red/color"
)

var (
	_ Material = &Standard{}
	_ Material = &BlinnPhong{}
)

type Material interface {
	Name() string
	Config(...Option)
}

type Standard struct {
	FlatShading      bool
	AmbientOcclusion bool
	ReceiveShadow    bool
	Texture          *buffer.Texture

	name string
}

func (m *Standard) Name() string {
	if m.name == "" {
		return "standard"
	}
	return m.name
}

func (m *Standard) Config(opts ...Option) {
	for _, opt := range opts {
		opt(m)
	}
}

type BlinnPhong struct {
	Standard
	Ambient   color.RGBA
	Diffuse   color.RGBA
	Specular  color.RGBA
	Emissive  color.RGBA
	Shininess float32
	Opacity   float32
}

// NewBlinnPhong creates a new Blinn-Phong material and returns the material ID.
// To configure the material, one can use Get to fetch the Material then use
// Config methods to customize the material.
func NewBlinnPhong(opts ...Option) ID {
	t := &BlinnPhong{
		Standard: Standard{
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
	id, _ := Put(t)
	return id
}

func (m *BlinnPhong) Name() string {
	if m.name == "" {
		return "blinn_phong"
	}
	return m.name
}

func (m *BlinnPhong) Config(opts ...Option) {
	for _, opt := range opts {
		opt(m)
	}
}
