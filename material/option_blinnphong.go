// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
)

type BlinnPhongOption func(m *BlinnPhong)

func Texture(tex *buffer.Texture) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.Texture = tex
	}
}

func Kdiff(col color.RGBA) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.Diffuse = col
	}
}

func Kspec(col color.RGBA) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.Specular = col
	}
}

func Shininess(val float32) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.Shininess = val
	}
}

func FlatShading(enable bool) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.FlatShading = true
	}
}

func AmbientOcclusion(enable bool) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.AmbientOcclusion = true
	}
}

func ReceiveShadow(enable bool) BlinnPhongOption {
	return func(m *BlinnPhong) {
		m.ReceiveShadow = true
	}
}
