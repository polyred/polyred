// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package term

import (
	"fmt"
	"image"
	"io"
)

// Terminal is a drawing canvas on system terminal.
type Terminal struct {
	buf           []byte
	width, height int
}

// Opt represents a terminal option.
type Opt func(t *Terminal)

// Size is an option to configure terminal drawing size.
func Size(w, h int) Opt {
	return func(t *Terminal) {
		t.width = w
		t.height = h
	}
}

// New returns a terminal instance.
func New(opts ...Opt) *Terminal {
	t := &Terminal{}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

var (
	// https://en.wikipedia.org/wiki/Block_Elements
	pixChar  = []byte("█")
	pixBreak = []byte("\n")
)

// Draw draws a given image on the terminal canvas.
func (t *Terminal) Draw(img *image.RGBA) {
	for j := 0; j < t.height; j++ {
		for i := 0; i < t.width; i++ {
			r, g, b, a := img.At(i, j).RGBA()
			switch {
			case a == 0:
				t.buf = append(t.buf, fgBytes([]byte(" "), uint8(r), uint8(g), uint8(b))...)
				continue
			case a < 255:
				alpha := float64(uint8(a)) / 255
				r = uint32(float64(r) * alpha)
				g = uint32(float64(g) * alpha)
				b = uint32(float64(b) * alpha)
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
func (t *Terminal) Flush(w io.Writer) {
	fmt.Fprint(w, string(t.buf))
}
