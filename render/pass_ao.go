// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/material"
	"poly.red/math"
	"poly.red/texture/buffer"
)

type ambientOcclusionPass struct{ buf *buffer.Buffer }

func (ao *ambientOcclusionPass) Shade(x, y int, col color.RGBA) color.RGBA {
	// FIXME: naive and super slow SSAO implementation. Optimize
	// when denoiser is available.

	info := ao.buf.At(x, y)
	mat, ok := info.AttrFlat["Mat"].(material.Material)
	if !ok {
		mat = nil
	}
	if mat == nil || !mat.AmbientOcclusion() {
		return col
	}

	total := float32(0.0)
	for a := float32(0.0); a < math.TwoPi-1e-4; a += math.Pi / 4 {
		total += math.HalfPi - ao.maxElevationAngle(x, y, math.Cos(a), math.Sin(a))
	}
	total /= (math.Pi / 2) * 8
	total = math.Pow(total, 10000)

	return color.RGBA{
		uint8(total * float32(col.R)),
		uint8(total * float32(col.G)),
		uint8(total * float32(col.B)), col.A}
}

func (ao *ambientOcclusionPass) maxElevationAngle(x, y int, dirX, dirY float32) float32 {
	p := math.NewVec4(float32(x), float32(y), 0, 1)
	dir := math.NewVec4(dirX, dirY, 0, 0)
	maxangle := float32(0.0)
	for t := float32(0.0); t < 100; t += 1 {
		cur := p.Add(dir.Scale(t, t, 1, 1))
		if !ao.buf.In(int(cur.X), int(cur.Y)) {
			return maxangle
		}

		distance := p.Sub(cur).Len()
		if distance < 1 {
			continue
		}

		// FIXME: I think the implementation here has internal bugs.
		// The minimum depth is assumed to be -1, otherwise the calculation
		// can be wrong. Figure out why.
		shadeInfo := ao.buf.At(int(cur.X), int(cur.Y))
		traceInfo := ao.buf.At(int(p.X), int(p.Y))
		shadeDepth := shadeInfo.Depth
		traceDepth := traceInfo.Depth
		if !shadeInfo.Ok {
			shadeDepth = -1
		}
		if !traceInfo.Ok {
			traceDepth = -1
		}
		elevation := shadeDepth - traceDepth
		maxangle = math.Max(maxangle, math.Atan(elevation/distance))
	}

	return maxangle
}
