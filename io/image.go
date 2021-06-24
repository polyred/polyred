// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package io

import (
	"fmt"
	"image"
	"image/draw"
	"os"
	"runtime"

	_ "image/jpeg"
	_ "image/png"

	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/utils"
)

// ImageOption offers several basic custom conversions of the image.
type ImageOption struct {
	gammaCorrection bool
}

type ReadImageOption func(o *ImageOption)

func WithGammaCorrection(enable bool) ReadImageOption {
	return func(o *ImageOption) {
		o.gammaCorrection = enable
	}
}

// MustLoadImage loads a given file into a texture.
func MustLoadImage(path string, opts ...ReadImageOption) *image.RGBA {
	img, err := LoadImage(path, opts...)
	if err != nil {
		panic(err)
	}
	return img
}

func LoadImage(path string, opts ...ReadImageOption) (*image.RGBA, error) {
	option := &ImageOption{
		gammaCorrection: false,
	}
	for _, opt := range opts {
		opt(option)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("loader: cannot open file %s, err: %w", path, err)
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return nil, fmt.Errorf("loader: cannot load texture, path: %s, err: %w", path, err)
	}

	var data *image.RGBA
	if v, ok := img.(*image.RGBA); ok {
		data = v
	} else {
		data = image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()))
		draw.Draw(data, data.Bounds(), img, img.Bounds().Min, draw.Src)
	}
	// Gamma correction, assume input space in sRGB and converting it to linear.
	if option.gammaCorrection {
		pool := utils.NewWorkerPool(uint64(runtime.GOMAXPROCS(0)))
		batch := 1 << 12 // empirical
		length := len(data.Pix)
		batcheEnd := length / (4 * batch)
		pool.Add(uint64(batcheEnd) + 1)

		// All batches with equal sizes
		for i := 0; i < batcheEnd*(4*batch); i += 4 * batch {
			offset := i
			pool.Execute(func() {
				for j := 0; j < 4*batch; j += 4 {
					data.Pix[offset+j+0] = uint8(color.FromsRGB2Linear(float64(data.Pix[offset+j+0])/0xff)*0xff + 0.5)
					data.Pix[offset+j+1] = uint8(color.FromsRGB2Linear(float64(data.Pix[offset+j+1])/0xff)*0xff + 0.5)
					data.Pix[offset+j+2] = uint8(color.FromsRGB2Linear(float64(data.Pix[offset+j+2])/0xff)*0xff + 0.5)
				}
			})
		}
		pool.Execute(func() {
			for i := batcheEnd * (4 * batch); i < length; i += 4 {
				data.Pix[i+0] = uint8(color.FromsRGB2Linear(float64(data.Pix[i+0])/0xff)*0xff + 0.5)
				data.Pix[i+1] = uint8(color.FromsRGB2Linear(float64(data.Pix[i+1])/0xff)*0xff + 0.5)
				data.Pix[i+2] = uint8(color.FromsRGB2Linear(float64(data.Pix[i+2])/0xff)*0xff + 0.5)
			}
		})

		pool.Wait()
	}

	return data, nil
}
