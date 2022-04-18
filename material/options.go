// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
)

type Option func(m Material)

func Name(name string) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *Standard:
			x.name = name
		case *BlinnPhong:
			x.Standard.name = name
		default:
			panic("unsupported type")
		}
	}
}

func Texture(tex *buffer.Texture) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *Standard:
			x.Texture = tex
		case *BlinnPhong:
			x.Standard.Texture = tex
		default:
			panic("unsupported type")
		}
	}
}

func Diffuse(col color.RGBA) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *BlinnPhong:
			x.Diffuse = col
		default:
			panic("unsupported type")
		}
	}
}

func Specular(col color.RGBA) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *BlinnPhong:
			x.Specular = col
		default:
			panic("unsupported type")
		}
	}
}

func Shininess(shininess float32) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *BlinnPhong:
			x.Shininess = shininess
		default:
			panic("unsupported type")
		}
	}
}

func FlatShading(enable bool) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *Standard:
			x.FlatShading = enable
		case *BlinnPhong:
			x.Standard.FlatShading = enable
		default:
			panic("unsupported type")
		}
	}
}

func AmbientOcclusion(enable bool) Option {
	return func(m Material) {
		switch x := any(m).(type) {
		case *Standard:
			x.AmbientOcclusion = enable
		case *BlinnPhong:
			x.Standard.AmbientOcclusion = enable
		default:
			panic("unsupported type")
		}
	}
}

func ReceiveShadow(enable bool) Option {
	return func(m Material) {
		switch x := m.(type) {
		case *Standard:
			x.ReceiveShadow = enable
		case *BlinnPhong:
			x.Standard.ReceiveShadow = enable
		default:
			panic("unsupported type")
		}
	}
}
