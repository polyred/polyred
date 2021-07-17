// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/math"
)

type ambientOcclusionPass struct {
	w, h    int
	gbuffer []gInfo
}

func (ao *ambientOcclusionPass) Shade(x, y int, col color.RGBA) color.RGBA {
	// FIXME: naive and super slow SSAO implementation. Optimize
	// when denoiser is available.
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
	p := math.NewVec4(float64(x), float64(y), 0, 1)
	dir := math.NewVec4(dirX, dirY, 0, 0)
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
