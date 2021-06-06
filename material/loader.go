// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package material

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// MustLoad loads a given file into a texture.
func MustLoad(path string) *Texture {
	f, err := os.Open(path)
	if err != nil {
		panic(fmt.Errorf("loader: cannot open file %s, err: %v", path, err))
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		panic(fmt.Errorf("cannot load texture, path: %s, err: %v", path, err))
	}
	var data *image.RGBA
	if v, ok := img.(*image.RGBA); ok {
		data = v
	} else {
		data = image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
		draw.Draw(data, data.Bounds(), img, img.Bounds().Min, draw.Src)
	}

	return NewTexture(data, true)
}
