// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
)

type materials interface {
	Standard | BlinnPhong
}

type Option[T materials] func(m *T)

func Texture[T materials](tex *buffer.Texture) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *Standard:
			x.Texture = tex
		case *BlinnPhong:
			x.Standard.Texture = tex
		default:
			panic("unsupported type")
		}
	}
}

func Diffuse[T BlinnPhong](col color.RGBA) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *BlinnPhong:
			x.Diffuse = col
		default:
			panic("unsupported type")
		}
	}
}

func Specular[T BlinnPhong](col color.RGBA) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *BlinnPhong:
			x.Specular = col
		default:
			panic("unsupported type")
		}
	}
}

func Shininess[T BlinnPhong](shininess float32) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *BlinnPhong:
			x.Shininess = shininess
		default:
			panic("unsupported type")
		}
	}
}

func FlatShading[T materials](enable bool) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *Standard:
			x.FlatShading = enable
		case *BlinnPhong:
			x.Standard.FlatShading = enable
		default:
			panic("unsupported type")
		}
	}
}

func AmbientOcclusion[T materials](enable bool) Option[T] {
	return func(m *T) {
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

func ReceiveShadow[T materials](enable bool) Option[T] {
	return func(m *T) {
		switch x := any(m).(type) {
		case *Standard:
			x.ReceiveShadow = enable
		case *BlinnPhong:
			x.Standard.ReceiveShadow = enable
		default:
			panic("unsupported type")
		}
	}
}
