// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"

	"changkun.de/x/polyred/color"
)

type FragmentShader func(x, y int) color.RGBA

// ScreenPass traserval all pixels of the given image buffer
func (r *Renderer) ScreenPass(buf *image.RGBA, shade FragmentShader) {
	w := buf.Bounds().Dx()
	h := buf.Bounds().Dy()

	blockSize := int(r.concurrentSize)
	wsteps := w / blockSize
	hsteps := h / blockSize
	defer r.workerPool.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.workerPool.Add(1)
		r.workerPool.Execute(func() {
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					col := shade(x, y)
					idx := x + w*y
					r.lockBuf[idx].Lock()
					buf.Pix[4*idx+0] = col.R
					buf.Pix[4*idx+1] = col.G
					buf.Pix[4*idx+2] = col.B
					buf.Pix[4*idx+3] = col.A
					r.lockBuf[idx].Unlock()
				}
			}
		})
		return
	}

	r.workerPool.Add(uint64(wsteps*hsteps) + 2)
	for i := 0; i < wsteps*blockSize; i += blockSize {
		for j := 0; j < hsteps*blockSize; j += blockSize {
			ii := i
			jj := j
			r.workerPool.Execute(func() {
				for k := 0; k < blockSize; k++ {
					for l := 0; l < blockSize; l++ {
						x := ii + k
						y := jj + l

						col := shade(x, y)
						idx := x + w*y
						r.lockBuf[idx].Lock()
						buf.Pix[4*idx+0] = col.R
						buf.Pix[4*idx+1] = col.G
						buf.Pix[4*idx+2] = col.B
						buf.Pix[4*idx+3] = col.A
						r.lockBuf[idx].Unlock()
					}
				}
			})
		}
	}
	r.workerPool.Execute(func() {
		for x := wsteps * blockSize; x < w; x++ {
			for y := 0; y < hsteps*blockSize; y++ {
				col := shade(x, y)
				idx := x + w*y
				r.lockBuf[idx].Lock()
				buf.Pix[4*idx+0] = col.R
				buf.Pix[4*idx+1] = col.G
				buf.Pix[4*idx+2] = col.B
				buf.Pix[4*idx+3] = col.A
				r.lockBuf[idx].Unlock()
			}
		}
	})
	r.workerPool.Execute(func() {
		for x := 0; x < wsteps*blockSize; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				col := shade(x, y)
				idx := x + w*y
				r.lockBuf[idx].Lock()
				buf.Pix[4*idx+0] = col.R
				buf.Pix[4*idx+1] = col.G
				buf.Pix[4*idx+2] = col.B
				buf.Pix[4*idx+3] = col.A
				r.lockBuf[idx].Unlock()
			}
		}
		for x := wsteps * blockSize; x < w; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				col := shade(x, y)
				idx := x + w*y
				r.lockBuf[idx].Lock()
				buf.Pix[4*idx+0] = col.R
				buf.Pix[4*idx+1] = col.G
				buf.Pix[4*idx+2] = col.B
				buf.Pix[4*idx+3] = col.A
				r.lockBuf[idx].Unlock()
			}
		}
	})
}
