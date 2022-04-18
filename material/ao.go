// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package material

import (
	"image/color"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/math"
)

type AmbientOcclusionPass struct{ Buf *buffer.FragmentBuffer }

func AmbientOcclusionShade(buf *buffer.FragmentBuffer, info *primitive.Fragment) color.RGBA {
	// FIXME: naive and super slow SSAO implementation. Optimize
	// when denoiser is available.
	mat := Get(ID(info.MaterialID)).(*BlinnPhong)
	if mat == nil || !mat.AmbientOcclusion {
		return info.Col
	}

	total := float32(0.0)
	for a := float32(0.0); a < math.TwoPi-1e-4; a += math.Pi / 4 {
		total += math.HalfPi - maxElevationAngle(buf, info, math.Cos(a), math.Sin(a))
	}
	total /= (math.Pi / 2) * 8
	total = math.Pow(total, 10000)

	return color.RGBA{
		uint8(total * float32(info.Col.R)),
		uint8(total * float32(info.Col.G)),
		uint8(total * float32(info.Col.B)), info.Col.A}
}

func maxElevationAngle(buf *buffer.FragmentBuffer, info *primitive.Fragment, dirX, dirY float32) float32 {
	p := math.NewVec4(float32(info.X), float32(info.Y), 0, 1)
	dir := math.NewVec4(dirX, dirY, 0, 0)
	maxangle := float32(0.0)
	for t := float32(0.0); t < 100; t += 1 {
		cur := p.Add(dir.Scale(t, t, 1, 1))
		if !buf.In(int(cur.X), int(cur.Y)) {
			return maxangle
		}

		distance := p.Sub(cur).Len()
		if distance < 1 {
			continue
		}

		// FIXME: I think the implementation here has internal bugs.
		// The minimum depth is assumed to be -1, otherwise the calculation
		// can be wrong. Figure out why.
		shadeInfo := buf.Get(int(cur.X), int(cur.Y))
		traceInfo := buf.Get(int(p.X), int(p.Y))
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
