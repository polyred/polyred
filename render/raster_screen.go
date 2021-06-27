// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"

	"changkun.de/x/polyred/color"
	"changkun.de/x/polyred/math"
)

// FragmentShader is a shader that executes on a frame buffer.
// The x and y are the row and column of the current executing pixel,
// and the col is the original color of the pixel at the frame buffer.
type FragmentShader func(x, y int, col color.RGBA) color.RGBA

// ScreenPass is a concurrent executor of the given shader that travel
// through all pixels. Each pixel executes the given shader exactly once.
// One should not manipulate the given image buffer in the shader. Instead,
// return the resulting color in the shader can avoid data race.
func (r *Renderer) ScreenPass(buf *image.RGBA, shade FragmentShader) {
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

	defer r.workerPool.Wait()

	if wsteps == 0 && hsteps == 0 {
		r.workerPool.Add(1)

		// Note: sadly that the executing function will escape to the
		// heap which increases the memory allocation. No workaround.
		r.workerPool.Execute(func() {
			for x := 0; x < w; x++ {
				for y := 0; y < h; y++ {
					old := buf.RGBAAt(x, y)
					col := shade(x, y, old)
					if col != color.Discard {
						buf.Set(x, y, col)
					}
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
						old := buf.RGBAAt(x, y)
						col := shade(x, y, old)
						if col != color.Discard {
							buf.Set(x, y, col)
						}
					}
				}
			})
		}
	}

	r.workerPool.Execute(func() {
		for x := wsteps * blockSize; x < w; x++ {
			for y := 0; y < hsteps*blockSize; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(x, y, old)
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
	}, func() {
		for x := 0; x < wsteps*blockSize; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(x, y, old)
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
		for x := wsteps * blockSize; x < w; x++ {
			for y := hsteps * blockSize; y < h; y++ {
				old := buf.RGBAAt(x, y)
				col := shade(x, y, old)
				if col != color.Discard {
					buf.Set(x, y, col)
				}
			}
		}
	})
}

type ambientOcclusionPass struct {
	w, h    int
	gbuffer []gInfo
}

func (ao *ambientOcclusionPass) Shade(x, y int, col color.RGBA) color.RGBA {
	// FIXME: naive and super slow SSAO implementation. Optimize
	// when denoiser is avaliable.
	w := ao.w
	idx := x + w*y
	info := &ao.gbuffer[idx]
	if info.mat == nil {
		return col
	}
	if !info.mat.AmbientOcclusion() {
		return col
	}

	total := 0.0
	for a := 0.0; a < math.Pi*2-1e-4; a += math.Pi / 4 {
		total += math.Pi/2 - ao.maxElevationAngle(x, y, math.Cos(a), math.Sin(a))
	}
	total /= (math.Pi / 2) * 8
	total = math.Pow(total, 10000)

	return color.RGBA{
		uint8(total * float64(col.R)),
		uint8(total * float64(col.G)),
		uint8(total * float64(col.B)), col.A}
}

func (ao *ambientOcclusionPass) maxElevationAngle(x, y int, dirX, dirY float64) float64 {
	p := math.NewVector(float64(x), float64(y), 0, 1)
	dir := math.NewVector(dirX, dirY, 0, 0)
	maxangle := 0.0
	for t := 0.0; t < 100; t += 1 {
		cur := p.Add(dir.Scale(t, t, 1, 1))
		if cur.X >= float64(ao.w) || cur.Y >= float64(ao.h) || cur.X < 0 || cur.Y < 0 {
			return maxangle
		}

		distance := p.Sub(cur).Len()
		if distance < 1 {
			continue
		}
		shadeIdx := int(cur.X) + ao.w*int(cur.Y)
		traceIdx := int(p.X) + ao.w*int(p.Y)

		elevation := ao.gbuffer[shadeIdx].z - ao.gbuffer[traceIdx].z
		maxangle = math.Max(maxangle, math.Atan(elevation/distance))
	}
	return maxangle
}
