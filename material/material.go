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

// NewBlinnPhong creates and returns a new Blinn-Phong material. Materials are
// owned by the geometry that uses them (geometry.New) and tabulated per render by
// the renderer; there is no global pool. Use Config to customize after creation.
func NewBlinnPhong(opts ...Option) *BlinnPhong {
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
	return t
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
