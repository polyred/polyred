// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"

	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/math"
)

type plane struct {
	pos, nor math.Vector
}

func (p plane) pointInFront(v math.Vector) bool {
	return v.Sub(p.pos).Dot(p.nor) > 0
}

func (p plane) intersectSegment(v0, v1 math.Vector) math.Vector {
	u := v1.Sub(v0)
	w := v0.Sub(p.pos)
	d := p.nor.Dot(u)
	n := -p.nor.Dot(w)
	s := n / d
	return v0.Add(u.Scale(s, s, s, s))
}

func sutherlandHodgman(points []math.Vector, w, h float64) []math.Vector {
	planes := []plane{
		{math.NewVector(w, 0, 0, 1), math.NewVector(-1, 0, 0, 1)},
		{math.NewVector(0, 0, 0, 1), math.NewVector(1, 0, 0, 1)},
		{math.NewVector(0, h, 0, 1), math.NewVector(0, -1, 0, 1)},
		{math.NewVector(0, 0, 0, 1), math.NewVector(0, 1, 0, 1)},
		{math.NewVector(0, 0, 1, 1), math.NewVector(0, 0, -1, 1)},
		{math.NewVector(0, 0, -1, 1), math.NewVector(0, 0, 1, 1)},
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

func (r *Renderer) clipTriangle(v1, v2, v3 *primitive.Vertex, w, h float64) []*primitive.Triangle {
	w1 := v1.Pos
	w2 := v2.Pos
	w3 := v3.Pos
	p1 := w1.Vec()
	p2 := w2.Vec()
	p3 := w3.Vec()
	points := []math.Vector{w1, w2, w3}
	newPoints := sutherlandHodgman(points, w, h)
	var result []*primitive.Triangle
	for i := 2; i < len(newPoints); i++ {
		b1w1, b1w2, b1w3 := math.Barycoord(newPoints[0], p1, p2, p3)
		b2w1, b2w2, b2w3 := math.Barycoord(newPoints[i-1], p1, p2, p3)
		b3w1, b3w2, b3w3 := math.Barycoord(newPoints[i], p1, p2, p3)

		// FIXME: Perspective corrected interpolation?

		t1 := primitive.Vertex{
			Pos: math.Vector{
				X: b1w1*v1.Pos.X + b1w2*v2.Pos.X + b1w3*v3.Pos.X,
				Y: b1w1*v1.Pos.Y + b1w2*v2.Pos.Y + b1w3*v3.Pos.Y,
				Z: b1w1*v1.Pos.Z + b1w2*v2.Pos.Z + b1w3*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vector{
				X: b1w1*v1.UV.X + b1w2*v2.UV.X + b1w3*v3.UV.X,
				Y: b1w1*v1.UV.Y + b1w2*v2.UV.Y + b1w3*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vector{
				X: b1w1*v1.Nor.X + b1w2*v2.Nor.X + b1w3*v3.Nor.X,
				Y: b1w1*v1.Nor.Y + b1w2*v2.Nor.Y + b1w3*v3.Nor.Y,
				Z: b1w1*v1.Nor.Z + b1w2*v2.Nor.Z + b1w3*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b1w1*float64(v1.Col.R)+b1w2*float64(v2.Col.R)+b1w3*float64(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b1w1*float64(v1.Col.G)+b1w2*float64(v2.Col.G)+b1w3*float64(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b1w1*float64(v1.Col.B)+b1w2*float64(v2.Col.B)+b1w3*float64(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b1w1*float64(v1.Col.A)+b1w2*float64(v2.Col.A)+b1w3*float64(v3.Col.A), 0, 0xff)),
			},
		}
		t2 := primitive.Vertex{
			Pos: math.Vector{
				X: b2w1*v1.Pos.X + b2w2*v2.Pos.X + b2w3*v3.Pos.X,
				Y: b2w1*v1.Pos.Y + b2w2*v2.Pos.Y + b2w3*v3.Pos.Y,
				Z: b2w1*v1.Pos.Z + b2w2*v2.Pos.Z + b2w3*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vector{
				X: b2w1*v1.UV.X + b2w2*v2.UV.X + b2w3*v3.UV.X,
				Y: b2w1*v1.UV.Y + b2w2*v2.UV.Y + b2w3*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vector{
				X: b2w1*v1.Nor.X + b2w2*v2.Nor.X + b2w3*v3.Nor.X,
				Y: b2w1*v1.Nor.Y + b2w2*v2.Nor.Y + b2w3*v3.Nor.Y,
				Z: b2w1*v1.Nor.Z + b2w2*v2.Nor.Z + b2w3*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b2w1*float64(v1.Col.R)+b2w2*float64(v2.Col.R)+b2w3*float64(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b2w1*float64(v1.Col.G)+b2w2*float64(v2.Col.G)+b2w3*float64(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b2w1*float64(v1.Col.B)+b2w2*float64(v2.Col.B)+b2w3*float64(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b2w1*float64(v1.Col.A)+b2w2*float64(v2.Col.A)+b2w3*float64(v3.Col.A), 0, 0xff)),
			},
		}
		t3 := primitive.Vertex{
			Pos: math.Vector{
				X: b3w1*v1.Pos.X + b3w2*v2.Pos.X + b3w3*v3.Pos.X,
				Y: b3w1*v1.Pos.Y + b3w2*v2.Pos.Y + b3w3*v3.Pos.Y,
				Z: b3w1*v1.Pos.Z + b3w2*v2.Pos.Z + b3w3*v3.Pos.Z,
				W: 1,
			},
			UV: math.Vector{
				X: b3w1*v1.UV.X + b3w2*v2.UV.X + b3w3*v3.UV.X,
				Y: b3w1*v1.UV.Y + b3w2*v2.UV.Y + b3w3*v3.UV.Y,
				Z: 0,
				W: 1,
			},
			Nor: math.Vector{
				X: b3w1*v1.Nor.X + b3w2*v2.Nor.X + b3w3*v3.Nor.X,
				Y: b3w1*v1.Nor.Y + b3w2*v2.Nor.Y + b3w3*v3.Nor.Y,
				Z: b3w1*v1.Nor.Z + b3w2*v2.Nor.Z + b3w3*v3.Nor.Z,
				W: 0,
			},
			Col: color.RGBA{
				R: uint8(math.Clamp(b3w1*float64(v1.Col.R)+b3w2*float64(v2.Col.R)+b3w3*float64(v3.Col.R), 0, 0xff)),
				G: uint8(math.Clamp(b3w1*float64(v1.Col.G)+b3w2*float64(v2.Col.G)+b3w3*float64(v3.Col.G), 0, 0xff)),
				B: uint8(math.Clamp(b3w1*float64(v1.Col.B)+b3w2*float64(v2.Col.B)+b3w3*float64(v3.Col.B), 0, 0xff)),
				A: uint8(math.Clamp(b3w1*float64(v1.Col.A)+b3w2*float64(v2.Col.A)+b3w3*float64(v3.Col.A), 0, 0xff)),
			},
		}
		result = append(result, primitive.NewTriangle(&t1, &t2, &t3))
	}
	return result
}
