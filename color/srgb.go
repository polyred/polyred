// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package color

import (
	"sync"

	"poly.red/math"
)

// FromLinear2sRGB converts a given value from linear space to
// sRGB space.
func FromLinear2sRGB(v float32) float32 {
	if !useLut {
		return linear2sRGB(v)
	}
	if v <= 0 {
		return 0
	}
	if v == 1 {
		return 1
	}
	i := v * lutSize
	ifloor := int(i) & (lutSize - 1)
	v0 := lin2sRGBLUT[ifloor]
	v1 := lin2sRGBLUT[ifloor+1]
	i -= float32(ifloor)
	return v0*(1.0-i) + v1*i
}

// FromsRGB2Linear converts a given value from linear space to
// sRGB space.
func FromsRGB2Linear(v float32) float32 {
	if !useLut {
		return sRGB2linear(v)
	}
	if v <= 0 {
		return 0
	}
	if v >= 1 {
		return 1
	}

	i := v * lutSize
	ifloor := int(i) & (lutSize - 1)
	v0 := sRGB2linLUT[ifloor]
	v1 := sRGB2linLUT[ifloor+1]
	i -= float32(ifloor)
	return v0*(1.0-i) + v1*i
}

var once sync.Once

const (
	lutSize = 1024 // keep a power of 2
)

var (
	useLut      = true
	lin2sRGBLUT [lutSize + 1]float32
	sRGB2linLUT [lutSize + 1]float32
)

func init() {
	once.Do(func() {
		for i := 0; i < lutSize; i++ {
			lin2sRGBLUT[i] = linear2sRGB(float32(i) / lutSize)
			sRGB2linLUT[i] = sRGB2linear(float32(i) / lutSize)
		}
		lin2sRGBLUT[lutSize] = lin2sRGBLUT[lutSize-1]
		sRGB2linLUT[lutSize] = sRGB2linLUT[lutSize-1]
	})
}

func sRGB2linear(v float32) float32 {
	if v <= 0.04045 {
		v /= 12.92
	} else {
		v = math.Pow((v+0.055)/1.055, 2.4)
	}
	return v
}

func linear2sRGB(v float32) float32 {
	if v <= 0.0031308 {
		v *= 12.92
	} else {
		v = 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	return v
}
