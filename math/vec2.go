// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

import "math"

type Vec2 struct {
	X, Y float64
}

func NewVec2(x, y float64) Vec2 {
	return Vec2{x, y}
}

func (v Vec2) Eq(u Vec2) bool {
	if ApproxEq(v.X, v.X, Epsilon) &&
		ApproxEq(v.Y, v.Y, Epsilon) {
		return true
	}
	return false
}

func (v Vec2) Add(u Vec2) Vec2 {
	return Vec2{v.X + u.X, v.Y + u.Y}
}

func (v Vec2) Sub(u Vec2) Vec2 {
	return Vec2{v.X - u.X, v.Y - u.Y}
}

func (v Vec2) IsZero() bool {
	if ApproxEq(v.X, 0, Epsilon) &&
		ApproxEq(v.Y, 0, Epsilon) {
		return true
	}
	return false
}

func (v Vec2) Scale(x, y float64) Vec2 {
	return Vec2{v.X * x, v.Y * y}
}

func (v Vec2) Translate(x, y float64) Vec2 {
	return Vec2{v.X + x, v.Y + y}
}

func (v Vec2) Dot(u Vec2) float64 {
	return v.X*u.X + v.Y*u.Y
}

func (v Vec2) Len() float64 {
	return math.Sqrt(v.Dot(v))
}

func (v Vec2) Unit() Vec2 {
	n := 1.0 / v.Len()
	return Vec2{v.X * n, v.Y * n}
}

func (v Vec2) Apply(m Mat2) Vec2 {
	x := m.X00*v.X + m.X01*v.Y
	y := m.X10*v.X + m.X11*v.Y
	return Vec2{x, y}
}
