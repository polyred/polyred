package texture

import (
	"image"

	"poly.red/color"
)

type TextureOption func(t *textureOption)

type textureOption struct {
	format Format
	img    *image.RGBA
	debug  bool
}

func WithFormat(format Format) TextureOption {
	return func(t *textureOption) {
		t.format = format
	}
}

func WithImage(img *image.RGBA) TextureOption {
	return func(t *textureOption) {
		if img.Bounds().Dx() < 1 || img.Bounds().Dy() < 1 {
			panic("image width or height is less than 1!")
		}
		t.img = img
	}
}

func WithColor(c color.RGBA) TextureOption {
	return func(t *textureOption) {
		img := &image.RGBA{
			Pix:    []uint8{c.R, c.G, c.B, c.A},
			Stride: 4,
			Rect:   image.Rect(0, 0, 1, 1),
		}

		if img.Bounds().Dx() < 1 || img.Bounds().Dy() < 1 {
			panic("image width or height is less than 1!")
		}

		t.img = img
	}
}

func WithDebug(enable bool) TextureOption {
	return func(t *textureOption) {
		t.debug = enable
	}
}
