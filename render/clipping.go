// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"poly.red/geometry/primitive"
	"poly.red/math"
)

type plane struct {
	pos, nor math.Vec4
}

func (p plane) pointInFront(v math.Vec4) bool {
	return v.Sub(p.pos).Dot(p.nor) > 0
}

func (p plane) intersectSegment(v0, v1 math.Vec4) math.Vec4 {
	u := v1.Sub(v0)
	w := v0.Sub(p.pos)
	d := p.nor.Dot(u)
	n := -p.nor.Dot(w)
	s := n / d
	return v0.Add(u.Scale(s, s, s, s))
}

func sutherlandHodgman(points []math.Vec4, w, h float32) []math.Vec4 {
	planes := []plane{
		{math.NewVec4(w, 0, 0, 1), math.NewVec4(-1, 0, 0, 1)},
		{math.NewVec4(0, 0, 0, 1), math.NewVec4(1, 0, 0, 1)},
		{math.NewVec4(0, h, 0, 1), math.NewVec4(0, -1, 0, 1)},
		{math.NewVec4(0, 0, 0, 1), math.NewVec4(0, 1, 0, 1)},
		{math.NewVec4(0, 0, 1, 1), math.NewVec4(0, 0, -1, 1)},
		{math.NewVec4(0, 0, -1, 1), math.NewVec4(0, 0, 1, 1)},
	}

	output := points
	for _, plane := range planes {
		input := output
		output = nil
		if len(input) == 0 {
			return nil
		}

		s := input[len(input)-1]
		for _, e := range input {
			if plane.pointInFront(e) {
				if !plane.pointInFront(s) {
					x := plane.intersectSegment(s, e)
					output = append(output, x)
				}
				output = append(output, e)
			} else if plane.pointInFront(s) {
				x := plane.intersectSegment(s, e)
				output = append(output, x)
			}
			s = e
		}
	}
	return output
}

func (r *Renderer) clipTriangle(v1, v2, v3 *primitive.Vertex, w, h float32, recipw math.Vec4) []*primitive.Triangle {
	p1 := v1.Pos
	p2 := v2.Pos
	p3 := v3.Pos
	clips := sutherlandHodgman([]math.Vec4{p1, p2, p3}, w, h)
	var result []*primitive.Triangle
	for i := 2; i < len(clips); i++ {
		// FIXME: clipping should be perspective correct
		b1bc := math.Barycoord(math.NewVec2(clips[0].X, clips[0].Y),
			p1.ToVec2(), p2.ToVec2(), p3.ToVec2())
		b2bc := math.Barycoord(math.NewVec2(clips[i-1].X, clips[i-1].Y),
			p1.ToVec2(), p2.ToVec2(), p3.ToVec2())
		b3bc := math.Barycoord(math.NewVec2(clips[i].X, clips[i].Y),
			p1.ToVec2(), p2.ToVec2(), p3.ToVec2())

		t1 := primitive.Vertex{
			Pos: math.Vec4{
				X: b1bc[0]*v1.Pos.X + b1bc[1]*v2.Pos.X + b1bc[2]*v3.Pos.X,
				Y: b1bc[0]*v1.Pos.Y + b1bc[1]*v2.Pos.Y + b1bc[2]*v3.Pos.Y,
				Z: b1bc[0]*v1.Pos.Z + b1bc[1]*v2.Pos.Z + b1bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b1bc[0]*v1.UV.X + b1bc[1]*v2.UV.X + b1bc[2]*v3.UV.X,
				Y: b1bc[0]*v1.UV.Y + b1bc[1]*v2.UV.Y + b1bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b1bc[0]*v1.Nor.X + b1bc[1]*v2.Nor.X + b1bc[2]*v3.Nor.X,
				Y: b1bc[0]*v1.Nor.Y + b1bc[1]*v2.Nor.Y + b1bc[2]*v3.Nor.Y,
				Z: b1bc[0]*v1.Nor.Z + b1bc[1]*v2.Nor.Z + b1bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b1bc[0]*float32(v1.Col.R)+b1bc[1]*float32(v2.Col.R)+b1bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b1bc[0]*float32(v1.Col.G)+b1bc[1]*float32(v2.Col.G)+b1bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b1bc[0]*float32(v1.Col.B)+b1bc[1]*float32(v2.Col.B)+b1bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b1bc[0]*float32(v1.Col.A)+b1bc[1]*float32(v2.Col.A)+b1bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}
		t2 := primitive.Vertex{
			Pos: math.Vec4{
				X: b2bc[0]*v1.Pos.X + b2bc[1]*v2.Pos.X + b2bc[2]*v3.Pos.X,
				Y: b2bc[0]*v1.Pos.Y + b2bc[1]*v2.Pos.Y + b2bc[2]*v3.Pos.Y,
				Z: b2bc[0]*v1.Pos.Z + b2bc[1]*v2.Pos.Z + b2bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b2bc[0]*v1.UV.X + b2bc[1]*v2.UV.X + b2bc[2]*v3.UV.X,
				Y: b2bc[0]*v1.UV.Y + b2bc[1]*v2.UV.Y + b2bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b2bc[0]*v1.Nor.X + b2bc[1]*v2.Nor.X + b2bc[2]*v3.Nor.X,
				Y: b2bc[0]*v1.Nor.Y + b2bc[1]*v2.Nor.Y + b2bc[2]*v3.Nor.Y,
				Z: b2bc[0]*v1.Nor.Z + b2bc[1]*v2.Nor.Z + b2bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b2bc[0]*float32(v1.Col.R)+b2bc[1]*float32(v2.Col.R)+b2bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b2bc[0]*float32(v1.Col.G)+b2bc[1]*float32(v2.Col.G)+b2bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b2bc[0]*float32(v1.Col.B)+b2bc[1]*float32(v2.Col.B)+b2bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b2bc[0]*float32(v1.Col.A)+b2bc[1]*float32(v2.Col.A)+b2bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}
		t3 := primitive.Vertex{
			Pos: math.Vec4{
				X: b3bc[0]*v1.Pos.X + b3bc[1]*v2.Pos.X + b3bc[2]*v3.Pos.X,
				Y: b3bc[0]*v1.Pos.Y + b3bc[1]*v2.Pos.Y + b3bc[2]*v3.Pos.Y,
				Z: b3bc[0]*v1.Pos.Z + b3bc[1]*v2.Pos.Z + b3bc[2]*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vec4{
				X: b3bc[0]*v1.UV.X + b3bc[1]*v2.UV.X + b3bc[2]*v3.UV.X,
				Y: b3bc[0]*v1.UV.Y + b3bc[1]*v2.UV.Y + b3bc[2]*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vec4{
				X: b3bc[0]*v1.Nor.X + b3bc[1]*v2.Nor.X + b3bc[2]*v3.Nor.X,
				Y: b3bc[0]*v1.Nor.Y + b3bc[1]*v2.Nor.Y + b3bc[2]*v3.Nor.Y,
				Z: b3bc[0]*v1.Nor.Z + b3bc[1]*v2.Nor.Z + b3bc[2]*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b3bc[0]*float32(v1.Col.R)+b3bc[1]*float32(v2.Col.R)+b3bc[2]*float32(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b3bc[0]*float32(v1.Col.G)+b3bc[1]*float32(v2.Col.G)+b3bc[2]*float32(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b3bc[0]*float32(v1.Col.B)+b3bc[1]*float32(v2.Col.B)+b3bc[2]*float32(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b3bc[0]*float32(v1.Col.A)+b3bc[1]*float32(v2.Col.A)+b3bc[2]*float32(v3.Col.A), 0, 0xff)),
			},
		}
		result = append(result, primitive.NewTriangle(&t1, &t2, &t3))
	}
	return result
}
