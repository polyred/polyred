// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package term

import (
	"fmt"
	"image"
	"image/draw"
	"io"
	"os"

	"poly.red/internal/imageutil"
)

// Terminal is a drawing canvas on system terminal.
type Terminal struct {
	buf           []byte
	out           io.Writer
	width, height int
}

// Option represents a terminal option.
type Option func(t *Terminal)

// Size is an option to configure terminal drawing size.
func Size(w, h int) Option {
	return func(t *Terminal) {
		t.width = w
		t.height = h
	}
}

// New returns a terminal instance.
func New(opts ...Option) *Terminal {
	t := &Terminal{
		out: os.Stdout,
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

var (
	// https://en.wikipedia.org/wiki/Block_Elements
	pixEmpty = []byte(" ")
	pixChar  = []byte("█")
	pixBreak = []byte("\n")
)

// Draw draws a given image on the terminal canvas.
// This draw call will resize and crop the source image to fit the
// target canvas.
func (t *Terminal) Draw(src *image.RGBA) {
	ftw := float64(t.width)
	fth := float64(t.height)

	// Crop and resize to fit the drawing canvas.
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	facW, facH := float64(w)/ftw, float64(h)/fth

	var dst *image.RGBA
	if dstH := fth * facW; dstH > float64(h) {
		dstW := w
		dst = image.NewRGBA(image.Rect(0, 0, dstW, int(dstH)))
	} else if dstW := ftw * facH; dstW > float64(h) {
		dstH := h
		dst = image.NewRGBA(image.Rect(0, 0, int(dstW), dstH))
	}
	draw.Draw(dst, dst.Bounds(), src, image.Pt(0, 0), draw.Over)

	// FIXME: should resize depends on the terminal pixel ratio.
	dst = imageutil.Resize(t.width*2, t.height, dst)

	for j := 0; j < t.height; j++ {
		for i := 0; i < t.width; i++ {
			r, g, b, a := dst.At(i, j).RGBA()
			switch {
			case a == 0:
				t.buf = append(t.buf, pixEmpty...)
				continue
			case a < 255:
				alpha := float64(uint8(a)) / 255
				fmt.Println(r, g, b)
				r = uint32(float64(r) * alpha)
				g = uint32(float64(g) * alpha)
				b = uint32(float64(b) * alpha)
				fmt.Println(r, g, b)
				fallthrough
			default:
				t.buf = append(t.buf, fgBytes(pixChar, uint8(r), uint8(g), uint8(b))...)
			}
		}
		t.buf = append(t.buf, pixBreak...)
	}
}

// Clear clears the existing terminal buffer.
// This function is not thread safe.
func (t *Terminal) Clear() {
	t.buf = t.buf[:0]
}

// Flush flushes the terminal buffer to the given io.Writer.
func (t *Terminal) Flush() {
	// TODO: support esapce sequence. We could flush it into terminal
	// and avoid the screen scrolling.

	fmt.Fprint(t.out, string(t.buf))
}
