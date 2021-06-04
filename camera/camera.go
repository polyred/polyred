// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package camera

import "changkun.de/x/ddd/math"

type Interface interface {
	Position() math.Vector
	ViewMatrix() math.Matrix
	ProjMatrix() math.Matrix
}

type PerspectiveCamera struct {
	position math.Vector
	lookAt   math.Vector
	up       math.Vector
	fov      float64
	aspect   float64
	near     float64
	far      float64
}

func NewPerspectiveCamera(pos, lookAt, up math.Vector, fov, aspect, near, far float64) PerspectiveCamera {
	return PerspectiveCamera{pos, lookAt, up, fov, aspect, near, far}
}

func (c PerspectiveCamera) Position() math.Vector {
	return c.position
}

func (c PerspectiveCamera) ViewMatrix() math.Matrix {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	Tr := math.NewMatrix(
		lxu.X, lxu.Y, lxu.Z, 0,
		u.X, u.Y, u.Z, 0,
		-l.X, -l.Y, -l.Z, 0,
		0, 0, 0, 1,
	)
	Tt := math.NewMatrix(
		1, 0, 0, -c.position.X,
		0, 1, 0, -c.position.Y,
		0, 0, 1, -c.position.Z,
		0, 0, 0, 1,
	)
	return Tr.MulM(Tt)
}

func (c PerspectiveCamera) ProjMatrix() math.Matrix {
	aspect := c.aspect
	fov := c.fov
	n := c.near
	f := c.far
	return math.NewMatrix(
		-1/(aspect*math.Tan((fov*math.Pi)/360)), 0, 0, 0,
		0, -1/math.Tan((fov*math.Pi)/360), 0, 0,
		0, 0, (n+f)/(n-f), (2*n*f)/(f-n),
		0, 0, 1, 0,
	)
}

type OrthographicCamera struct {
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

func NewOrthographicCamera(pos, lookAt, up math.Vector, fov, aspect, near, far float64) PerspectiveCamera {
	return PerspectiveCamera{pos, lookAt, up, fov, aspect, near, far}
}

func (c OrthographicCamera) Position() math.Vector {
	return c.position
}

func (c OrthographicCamera) ViewMatrix() math.Matrix {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	Tr := math.NewMatrix(
		lxu.X, lxu.Y, lxu.Z, 0,
		u.X, u.Y, u.Z, 0,
		-l.X, -l.Y, -l.Z, 0,
		0, 0, 0, 1,
	)
	Tt := math.NewMatrix(
		1, 0, 0, -c.position.X,
		0, 1, 0, -c.position.Y,
		0, 0, 1, -c.position.Z,
		0, 0, 0, 1,
	)
	return Tr.MulM(Tt)
}

func (c OrthographicCamera) ProjMatrix() math.Matrix {
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
