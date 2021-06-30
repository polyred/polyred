// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"

	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/shader"
)

// FIXME: use shader.Program

// ScreenPass is a concurrent executor of the given shader that travel
// through all pixels. Each pixel executes the given shader exactly once.
// One should not manipulate the given image buffer in the shader.
// Instead, return the resulting color in the shader can avoid data race.
func (r *Renderer) ScreenPass(buf *image.RGBA, shade shader.FragmentProgram) {
	if shade == nil {
		return
	}

	// Because the shader executes exactly on each pixel once, there is
	// no need to lock the pixel while reading or writing.

	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()

	blockSize := int(r.concurrentSize)
	wsteps := w / blockSize
	hsteps := h / blockSize

	defer r.sched.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.sched.Add(1)

		// Note: sadly that the executing function will escape to the
		// heap which increases the memory allocation. No workaround.
		r.sched.Execute(func() {
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					old := buf.RGBAAt(x, y)
					col := shade(primitive.Fragment{
						X: x, Y: y, Col: old,
					})
					if col != color.Discard {
						buf.Set(x, y, col)
					}
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
			r.sched.Execute(func() {
				for k := 0; k < blockSize; k++ {
					for l := 0; l < blockSize; l++ {
						x := ii + k
						y := jj + l
						old := buf.RGBAAt(x, y)
						col := shade(primitive.Fragment{
							X: x, Y: y, Col: old,
						})
						if col != color.Discard {
							buf.Set(x, y, col)
						}
					}
				}
			})
		}
	}

	r.sched.Execute(func() {
		for x := wsteps * blockSize; x < w; x++ {
			for y := 0; y < hsteps*blockSize; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
	}, func() {
		for x := 0; x < wsteps*blockSize; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
		for x := wsteps * blockSize; x < w; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(primitive.Fragment{
					X: x, Y: y, Col: old,
				})
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
	})
}
