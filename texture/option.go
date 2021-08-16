// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

import "image"

type Opt func(t interface{})

func Image(data *image.RGBA) Opt {
	return func(t interface{}) {
		switch o := t.(type) {
		case *Texture:
			if data.Bounds().Dx() < 1 || data.Bounds().Dy() < 1 {
				panic("image width or height is less than 1!")
			}
			o.image = data
		default:
			panic("texture: misuse of Image option")
		}
	}
}

func Debug(enable bool) Opt {
	return func(t interface{}) {
		switch o := t.(type) {
		case *Texture:
			o.debug = enable
		default:
			panic("texture: misuse of Debug option")
		}
	}
}

// IsoMipmap is a isotropic mipmap option
func IsoMipmap(enable bool) Opt {
	return func(t interface{}) {
		switch o := t.(type) {
		case *Texture:
			o.useMipmap = enable
		default:
			panic("texture: misuse of IsoMipmap option")
		}
	}
}
