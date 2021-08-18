// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"

	"poly.red/color"
	"poly.red/geometry/primitive"
	"poly.red/shader"
	"poly.red/texture/buffer"
)

// BlendFunc is a blending function for two given colors and returns
// the resulting color.
type BlendFunc func(dst, src color.RGBA) color.RGBA

// AlphaBlend performs alpha blending for pre-multiplied alpha RGBA colors
func AlphaBlend(dst, src color.RGBA) color.RGBA {
	// FIXME: there is an overflow
	sr, sg, sb, sa := uint32(src.R), uint32(src.G), uint32(src.B), uint32(src.A)
	dr, dg, db, da := uint32(dst.R), uint32(dst.G), uint32(dst.B), uint32(dst.A)

	// dr, dg, db and da are all 8-bit color at the moment, ranging in [0,255].
	// We work in 16-bit color, and so would normally do:
	// dr |= dr << 8
	// and similarly for dg, db and da, but instead we multiply a
	// (which is a 16-bit color, ranging in [0,65535]) by 0x101.
	// This yields the same result, but is fewer arithmetic operations.
	a := (0xffff - sa) * 0x101

	r := sr + dr*a/0xffff
	g := sg + dg*a/0xffff
	b := sb + db*a/0xffff
	aa := sa + da*a/0xffff
	return color.RGBA{uint8(r), uint8(g), uint8(b), uint8(aa)}
}

// DrawFragment is a concurrent executor of the given shader that travel
// through all pixels. Each pixel executes the given shader exactly once.
//
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) DrawFragments(buf *buffer.Buffer, shade shader.FragmentProgram) {
	r.DrawPixels(buf.Image(), shade)
}

// DrawPixels is a concurrent executor of the given shader that travel
// through all pixels. Each pixel executes the given shader exactly once.
//
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) DrawPixels(buf *image.RGBA, shade shader.FragmentProgram) {
	if shade == nil {
		return
	}

	// Because the shader executes exactly on each pixel once, there is
	// no need to lock the pixel while reading or writing.

	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()

	blockSize := int(r.batchSize)
	wsteps := w / blockSize
	hsteps := h / blockSize

	defer r.sched.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.sched.Add(1)

		// Note: sadly that the executing function will escape to the
		// heap which increases the memory allocation. No workaround.
		r.sched.Run(func() {
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {

					idx := buf.PixOffset(x, y)
					s := buf.Pix[idx : idx+4 : idx+4]
					old := color.RGBA{s[0], s[1], s[2], s[3]}

					// TODO: support multiple shaders.
					// The following code can impact performance significantly.
					// Figure out why.
					//
					// var col color.RGBA
					// for _, f := range shade {
					// 	col = f(primitive.Fragment{X: x, Y: y, Col: old})
					// 	if col == color.Discard {
					// 		break
					// 	}
					// }
					// if col == color.Discard {
					// 	continue
					// }
					col := shade(primitive.Fragment{
						X: x, Y: y, Col: old,
					})
					if col == color.Discard {
						continue
					}
					if r.blendFunc != nil {
						col = r.blendFunc(old, col)
					}
					// Use SetRGBA instead of Set can avoid memory allocation.
					//
					s[0] = col.R
					s[1] = col.G
					s[2] = col.B
					s[3] = col.A
				}
			}
		})
		return
	}

	r.sched.Add(uint64(wsteps*hsteps) + 2)
	for i := 0; i < wsteps*blockSize; i += blockSize {
		for j := 0; j < hsteps*blockSize; j += blockSize {
			ii := i
			jj := j
			r.sched.Run(func() {
				for k := 0; k < blockSize; k++ {
					for l := 0; l < blockSize; l++ {
						x := ii + k
						y := jj + l

						idx := buf.PixOffset(x, y)
						s := buf.Pix[idx : idx+4 : idx+4]
						old := color.RGBA{s[0], s[1], s[2], s[3]}

						col := shade(primitive.Fragment{
							X: x, Y: y, Col: old,
						})
						if col == color.Discard {
							continue
						}
						if r.blendFunc != nil {
							col = r.blendFunc(old, col)
						}
						// Use SetRGBA instead of Set can avoid memory allocation.
						// See https://golang.org/issue/44808.
						s[0] = col.R
						s[1] = col.G
						s[2] = col.B
						s[3] = col.A
					}
				}
			})
		}
	}

	r.sched.Run(func() {
		for x := wsteps * blockSize; x < w; x++ {
			for y := 0; y < hsteps*blockSize; y++ {

				idx := buf.PixOffset(x, y)
				s := buf.Pix[idx : idx+4 : idx+4]
				old := color.RGBA{s[0], s[1], s[2], s[3]}

				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col == color.Discard {
					continue
				}
				if r.blendFunc != nil {
					col = r.blendFunc(old, col)
				}
				// Use SetRGBA instead of Set can avoid memory allocation.
				// See https://golang.org/issue/44808.
				s[0] = col.R
				s[1] = col.G
				s[2] = col.B
				s[3] = col.A
			}
		}
	}, func() {
		for x := 0; x < wsteps*blockSize; x++ {
			for y := hsteps * blockSize; y < h; y++ {

				idx := buf.PixOffset(x, y)
				s := buf.Pix[idx : idx+4 : idx+4]
				old := color.RGBA{s[0], s[1], s[2], s[3]}

				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col == color.Discard {
					continue
				}
				if r.blendFunc != nil {
					col = r.blendFunc(old, col)
				}
				// Use SetRGBA instead of Set can avoid memory allocation.
				// See https://golang.org/issue/44808.
				s[0] = col.R
				s[1] = col.G
				s[2] = col.B
				s[3] = col.A
			}
		}
		for x := wsteps * blockSize; x < w; x++ {
			for y := hsteps * blockSize; y < h; y++ {

				idx := buf.PixOffset(x, y)
				s := buf.Pix[idx : idx+4 : idx+4]
				old := color.RGBA{s[0], s[1], s[2], s[3]}

				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col == color.Discard {
					continue
				}
				if r.blendFunc != nil {
					col = r.blendFunc(old, col)
				}
				// Use SetRGBA instead of Set can avoid memory allocation.
				// See https://golang.org/issue/44808.
				s[0] = col.R
				s[1] = col.G
				s[2] = col.B
				s[3] = col.A
			}
		}
	})
}
