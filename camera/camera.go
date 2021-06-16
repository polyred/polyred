// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import "changkun.de/x/ddd/math"

type Interface interface {
	Position() math.Vector
	ViewMatrix() math.Matrix
	ProjMatrix() math.Matrix
}

// ViewMatrix is a handy function for computing view matrix.
func ViewMatrix(pos, lookAt, up math.Vector) math.Matrix {
	l := lookAt.Sub(pos).Unit()
	lxu := l.Cross(up).Unit()
	u := lxu.Cross(l).Unit()
	x := pos.X
	y := pos.Y
	z := pos.Z
	// Tr := math.NewMatrix(
	// 	lxu.X, lxu.Y, lxu.Z, 0,
	// 	u.X, u.Y, u.Z, 0,
	// 	-l.X, -l.Y, -l.Z, 0,
	// 	0, 0, 0, 1,
	// )
	// Tt := math.NewMatrix(
	// 	1, 0, 0, -x,
	// 	0, 1, 0, -y,
	// 	0, 0, 1, -z,
	// 	0, 0, 0, 1,
	// )
	TrTt := math.NewMatrix(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt // Tr.MulM(Tt)
}

type Perspective struct {
	position math.Vector
	lookAt   math.Vector
	up       math.Vector
	fov      float64
	aspect   float64
	near     float64 // 0 < near < far
	far      float64
}

func NewPerspective(pos, lookAt, up math.Vector, fov, aspect, near, far float64) Perspective {
	return Perspective{pos, lookAt, up, fov, aspect, near, far}
}

func (c Perspective) Position() math.Vector {
	return c.position
}

func (c Perspective) ViewMatrix() math.Matrix {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	x := c.position.X
	y := c.position.Y
	z := c.position.Z
	// Tr := math.NewMatrix(
	// 	lxu.X, lxu.Y, lxu.Z, 0,
	// 	u.X, u.Y, u.Z, 0,
	// 	-l.X, -l.Y, -l.Z, 0,
	// 	0, 0, 0, 1,
	// )
	// Tt := math.NewMatrix(
	// 	1, 0, 0, -x,
	// 	0, 1, 0, -y,
	// 	0, 0, 1, -z,
	// 	0, 0, 0, 1,
	// )
	TrTt := math.NewMatrix(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt // Tr.MulM(Tt)
}

func (c Perspective) ProjMatrix() math.Matrix {
	aspect := c.aspect
	fov := (c.fov * math.Pi) / 180
	n := c.near
	f := c.far
	return math.NewMatrix(
		-1/(aspect*math.Tan(fov/2)), 0, 0, 0,
		0, -1/math.Tan(fov/2), 0, 0,
		0, 0, (n+f)/(n-f), (2*n*f)/(n-f),
		0, 0, 1, 0,
	)
}

type Orthographic struct {
	position math.Vector
	lookAt   math.Vector
	up       math.Vector
	left     float64
	right    float64
	bottom   float64
	top      float64
	near     float64
	far      float64
}

func NewOrthographic(
	pos, lookAt, up math.Vector,
	left, right, bottom, top, near, far float64,
) Orthographic {
	return Orthographic{
		position: pos,
		lookAt:   lookAt,
		up:       up,
		left:     left,
		right:    right,
		bottom:   bottom,
		top:      top,
		near:     near,
		far:      far,
	}
}

func (c Orthographic) Position() math.Vector {
	return c.position
}

func (c Orthographic) ViewMatrix() math.Matrix {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	x := c.position.X
	y := c.position.Y
	z := c.position.Z
	// Tr := math.NewMatrix(
	// 	lxu.X, lxu.Y, lxu.Z, 0,
	// 	u.X, u.Y, u.Z, 0,
	// 	-l.X, -l.Y, -l.Z, 0,
	// 	0, 0, 0, 1,
	// )
	// Tt := math.NewMatrix(
	// 	1, 0, 0, -x,
	// 	0, 1, 0, -y,
	// 	0, 0, 1, -z,
	// 	0, 0, 0, 1,
	// )
	TrTt := math.NewMatrix(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt // Tr.MulM(Tt)
}

func (c Orthographic) ProjMatrix() math.Matrix {
	l := c.left
	r := c.right
	t := c.top
	b := c.bottom
	n := c.near
	f := c.far
	return math.NewMatrix(
		2/(r-l), 0, 0, (l+r)/(l-r),
		0, 2/(t-b), 0, (b+t)/(b-t),
		0, 0, 2/(n-f), (f+n)/(f-n),
		0, 0, 0, 1,
	)
}
