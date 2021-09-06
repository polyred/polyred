// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import "poly.red/texture"

type Opt func(m Material)

func Texture(tex *texture.Texture) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.tex = tex
		default:
			panic("material: misuse of Texture option")
		}
	}
}

func Kdiff(val float32) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.kDiff = val
		default:
			panic("material: misuse of Kdiff option")
		}
	}
}

func Kspec(val float32) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.kSpec = val
		default:
			panic("material: misuse of Kspec option")
		}
	}
}

func Shininess(val float32) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.shininess = val
		default:
			panic("material: misuse of Shininess option")
		}
	}
}

func FlatShading(enable bool) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.flatShading = enable
		default:
			panic("material: misuse of FlatShading option")
		}
	}
}

func AmbientOcclusion(enable bool) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.ambientOcclusion = enable
		default:
			panic("material: misuse of AmbientOcclusion option")
		}
	}
}

func ReceiveShadow(enable bool) Opt {
	return func(m Material) {
		switch mat := m.(type) {
		case *BlinnPhongMaterial:
			mat.receiveShadow = enable
		default:
			panic("material: misuse of ReceiveShadow option")
		}
	}
}
