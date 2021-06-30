// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package camera provides a camera abstraction for perspective and
// orthographic camera and their utilities, such as viewing transformation
// matrices.
package camera

import (
	"changkun.de/x/polyred/math"
	"changkun.de/x/polyred/object"
)

// Interface assertion
var (
	_ Interface = &Orthographic{}
	_ Interface = &Perspective{}
)

// Interface is a camera interface that represents either orthographic
// or perspective camera.
type Interface interface {
	object.Object

	Position() math.Vec4
	ViewMatrix() math.Mat4
	ProjMatrix() math.Mat4
}

// ViewMatrix is a handy function for computing view matrix without
// instantiating all required camera parameters. The camera view matrix
// is determined via its position, look at position, and a up direction.
func ViewMatrix(pos, lookAt, up math.Vec4) math.Mat4 {
	if lookAt.W != 1 {
		panic("camera: misuse of ViewMatrix")
	}

	l := lookAt.Sub(pos).Unit()
	lxu := l.Cross(up).Unit()
	u := lxu.Cross(l).Unit()
	x := pos.X
	y := pos.Y
	z := pos.Z
	TrTt := math.NewMat4(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt
}

// Perspective prepresents a perspective camera.
type Perspective struct {
	math.TransformContext

	position math.Vec4
	lookAt   math.Vec4
	up       math.Vec4
	fov      float64
	aspect   float64
	near     float64 // 0 < near < far
	far      float64
}

// NewPerspective creates a new perspective camera with the provided
// camera parameters. Note that the lookAt parameter must be a position
// instead of direction (i.e. the w component of the vector must be 1).
func NewPerspective(pos, lookAt, up math.Vec4, fov, aspect, near, far float64) Interface {
	if lookAt.W != 1 {
		panic("camera: misuse of perspective camera")
	}

	c := &Perspective{
		position: pos, lookAt: lookAt, up: up,
		fov: fov, aspect: aspect, near: near, far: far,
	}
	c.ResetContext()
	return c
}

// Type returns the object type, i.e. object.TypeCamera
func (c *Perspective) Type() object.Type {
	return object.TypeCamera
}

// Position returns the position of the given perspective camera.
func (c *Perspective) Position() math.Vec4 {
	return c.position
}

// ViewMatrix returns the view matrix of the given camera.
func (c *Perspective) ViewMatrix() math.Mat4 {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	x := c.position.X
	y := c.position.Y
	z := c.position.Z
	TrTt := math.NewMat4(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt
}

// ProjMatrix returns the projection matrix of the given camera.
func (c *Perspective) ProjMatrix() math.Mat4 {
	aspect := c.aspect
	fov := (c.fov * math.Pi) / 180
	n := c.near
	f := c.far
	return math.NewMat4(
		-1/(aspect*math.Tan(fov/2)), 0, 0, 0,
		0, -1/math.Tan(fov/2), 0, 0,
		0, 0, (n+f)/(n-f), (2*n*f)/(n-f),
		0, 0, 1, 0,
	)
}

// Orthographic represents a orthographic camera.
type Orthographic struct {
	math.TransformContext

	position math.Vec4
	lookAt   math.Vec4
	up       math.Vec4
	left     float64
	right    float64
	bottom   float64
	top      float64
	near     float64
	far      float64
}

// NewOrthographic creates an orthographic camera with the provided
// camera parameters. Note that the lookAt parameter must be a position
// instead of direction (i.e. the w component of the vector must be 1).
func NewOrthographic(
	pos, lookAt, up math.Vec4,
	left, right, bottom, top, near, far float64,
) Interface {
	if lookAt.W != 1 {
		panic("camera: misuse of orthographic camera")
	}

	c := &Orthographic{
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
	c.ResetContext()
	return c
}

// Type returns the object type, i.e. object.TypeCamera
func (c *Orthographic) Type() object.Type {
	return object.TypeCamera
}

// Position returns the position of the given orthographic camera.
func (c *Orthographic) Position() math.Vec4 {
	return c.position
}

// ViewMatrix returns the view matrix of the given camera.
func (c *Orthographic) ViewMatrix() math.Mat4 {
	l := c.lookAt.Sub(c.position).Unit()
	lxu := l.Cross(c.up).Unit()
	u := lxu.Cross(l).Unit()
	x := c.position.X
	y := c.position.Y
	z := c.position.Z
	TrTt := math.NewMat4(
		lxu.X, lxu.Y, lxu.Z, -lxu.X*x-lxu.Y*y-lxu.Z*z,
		u.X, u.Y, u.Z, -u.X*x-u.Y*y-u.Z*z,
		-l.X, -l.Y, -l.Z, l.X*x+l.Y*y+l.Z*z,
		0, 0, 0, 1,
	)
	return TrTt
}

// ProjMatrix returns the projection matrix of the given camera.
func (c *Orthographic) ProjMatrix() math.Mat4 {
	l := c.left
	r := c.right
	t := c.top
	b := c.bottom
	n := c.near
	f := c.far
	return math.NewMat4(
		2/(r-l), 0, 0, (l+r)/(l-r),
		0, 2/(t-b), 0, (b+t)/(b-t),
		0, 0, 2/(n-f), (f+n)/(f-n),
		0, 0, 0, 1,
	)
}
