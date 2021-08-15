// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package texture

import "image"

type Opt func(t *Texture)

func Image(data *image.RGBA) Opt {
	return func(t *Texture) {
		if data.Bounds().Dx() < 1 || data.Bounds().Dy() < 1 {
			panic("image width or height is less than 1!")
		}
		t.image = data
	}
}

func Debug(enable bool) Opt {
	return func(t *Texture) {
		t.debug = enable
	}
}

// IsoMipmap is a isotropic mipmap option
func IsoMipmap(enable bool) Opt {
	return func(t *Texture) {
		t.useMipmap = enable
	}
}
