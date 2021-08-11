// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package utils

import (
	"image"
	"math"
	"runtime"
	"sync"
)

func linear(in float64) float64 {
	in = math.Abs(in)
	if in <= 1 {
		return 1 - in
	}
	return 0
}

// Resize scales an image to new width and height using bilinear interpolation.
func Resize(width, height int, img *image.RGBA) *image.RGBA {
	scaleX, scaleY := calcFactors(width, height, float64(img.Bounds().Dx()), float64(img.Bounds().Dy()))
	if width == 0 {
		width = int(0.7 + float64(img.Bounds().Dx())/scaleX)
	}
	if height == 0 {
		height = int(0.7 + float64(img.Bounds().Dy())/scaleY)
	}

	// Trivial case: return input image
	if int(width) == img.Bounds().Dx() && int(height) == img.Bounds().Dy() {
		return img
	}

	// Input image has no pixels
	if img.Bounds().Dx() <= 0 || img.Bounds().Dy() <= 0 {
		return img
	}

	cpus := runtime.GOMAXPROCS(0)
	wg := sync.WaitGroup{}

	// 8-bit precision
	temp := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dy(), int(width)))
	result := image.NewRGBA(image.Rect(0, 0, int(width), int(height)))

	// horizontal filter, results in transposed temporary image
	coeffs, offset, filterLength := createWeights8(temp.Bounds().Dy(), 2, scaleX, linear)
	wg.Add(cpus)
	for i := 0; i < cpus; i++ {
		slice := makeSlice(temp, i, cpus).(*image.RGBA)
		go func() {
			defer wg.Done()
			resizeRGBA(img, slice, scaleX, coeffs, offset, filterLength)
		}()
	}
	wg.Wait()

	// horizontal filter on transposed image, result is not transposed
	coeffs, offset, filterLength = createWeights8(result.Bounds().Dy(), 2, scaleY, linear)
	wg.Add(cpus)
	for i := 0; i < cpus; i++ {
		slice := makeSlice(result, i, cpus).(*image.RGBA)
		go func() {
			defer wg.Done()
			resizeRGBA(temp, slice, scaleY, coeffs, offset, filterLength)
		}()
	}
	wg.Wait()
	return result
}

func resizeRGBA(in *image.RGBA, out *image.RGBA, scale float64, coeffs []int16, offset []int, filterLength int) {
	newBounds := out.Bounds()
	maxX := in.Bounds().Dx() - 1

	for x := newBounds.Min.X; x < newBounds.Max.X; x++ {
		row := in.Pix[x*in.Stride:]
		for y := newBounds.Min.Y; y < newBounds.Max.Y; y++ {
			var rgba [4]int32
			var sum int32
			start := offset[y]
			ci := y * filterLength
			for i := 0; i < filterLength; i++ {
				coeff := coeffs[ci+i]
				if coeff != 0 {
					xi := start + i
					switch {
					case uint(xi) < uint(maxX):
						xi *= 4
					case xi >= maxX:
						xi = 4 * maxX
					default:
						xi = 0
					}

					rgba[0] += int32(coeff) * int32(row[xi+0])
					rgba[1] += int32(coeff) * int32(row[xi+1])
					rgba[2] += int32(coeff) * int32(row[xi+2])
					rgba[3] += int32(coeff) * int32(row[xi+3])
					sum += int32(coeff)
				}
			}

			xo := (y-newBounds.Min.Y)*out.Stride + (x-newBounds.Min.X)*4

			out.Pix[xo+0] = clampUint8(rgba[0] / sum)
			out.Pix[xo+1] = clampUint8(rgba[1] / sum)
			out.Pix[xo+2] = clampUint8(rgba[2] / sum)
			out.Pix[xo+3] = clampUint8(rgba[3] / sum)
		}
	}
}

// Calculates scaling factors using old and new image dimensions.
func calcFactors(width, height int, oldWidth, oldHeight float64) (scaleX, scaleY float64) {
	if width == 0 {
		if height == 0 {
			scaleX = 1.0
			scaleY = 1.0
		} else {
			scaleY = oldHeight / float64(height)
			scaleX = scaleY
		}
	} else {
		scaleX = oldWidth / float64(width)
		if height == 0 {
			scaleY = scaleX
		} else {
			scaleY = oldHeight / float64(height)
		}
	}
	return
}

type imageWithSubImage interface {
	image.Image
	SubImage(image.Rectangle) image.Image
}

func makeSlice(img imageWithSubImage, i, n int) image.Image {
	return img.SubImage(image.Rect(img.Bounds().Min.X, img.Bounds().Min.Y+i*img.Bounds().Dy()/n, img.Bounds().Max.X, img.Bounds().Min.Y+(i+1)*img.Bounds().Dy()/n))
}

// Keep value in [0,255] range.
func clampUint8(in int32) uint8 {
	// casting a negative int to an uint will result in an overflown
	// large uint. this behavior will be exploited here and in other functions
	// to achieve a higher performance.
	if uint32(in) < 256 {
		return uint8(in)
	}
	if in > 255 {
		return 255
	}
	return 0
}

// range [-256,256]
func createWeights8(dy, filterLength int, scale float64, kernel func(float64) float64) ([]int16, []int, int) {
	filterLength = filterLength * int(math.Max(math.Ceil(scale), 1))
	filterFactor := math.Min(1./(scale), 1)

	coeffs := make([]int16, dy*filterLength)
	start := make([]int, dy)
	for y := 0; y < dy; y++ {
		interpX := scale*(float64(y)+0.5) - 0.5
		start[y] = int(interpX) - filterLength/2 + 1
		interpX -= float64(start[y])
		for i := 0; i < filterLength; i++ {
			in := (interpX - float64(i)) * filterFactor
			coeffs[y*filterLength+i] = int16(kernel(in) * 256)
		}
	}

	return coeffs, start, filterLength
}
