// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package material

import (
	"fmt"
	"image"
	"image/png"
	"os"
)

// MustLoad loads a given file into a texture.
func MustLoad(path string) *Texture {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	img, err := png.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load texture, path: %s, err: %v", path, err))
	}
	var data *image.RGBA
	if v, ok := img.(*image.NRGBA); ok {
		data = (*image.RGBA)(v)
	} else if v, ok := img.(*image.RGBA); ok {
		data = v
	}

	return NewTexture(data)
}
