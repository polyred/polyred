// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import (
	"poly.red/math"
	"poly.red/object"
)

// Orthographic represents an orthographic camera.
type Orthographic struct {
	math.TransformContext

	position math.Vec3
	target   math.Vec3
	up       math.Vec3
	left     float32
	right    float32
	bottom   float32
	top      float32
	near     float32
	far      float32
}

// NewOrthographic creates an orthographic camera with the provided
// camera parameters.
func NewOrthographic(opts ...Opt) Interface {
	c := &Orthographic{
		position: math.NewVec3(0, 0, 1),
		target:   math.NewVec3(0, 0, 0),
		up:       math.NewVec3(0, 1, 0),
		left:     -1,
		right:    1,
		bottom:   -1,
		top:      1,
		near:     1,
		far:      -1,
	}
	for _, opt := range opts {
		opt(c)
	}

	c.ResetContext()
	return c
}

// Type returns the object type, i.e. object.TypeCamera
func (c *Orthographic) Type() object.Type {
	return object.TypeCamera
}

// Fov returns the field of view of the given camera
func (c *Orthographic) Fov() float32 {
	return 2 * math.Atan(c.top/math.Abs(c.near))
}

// Aspect returns the aspect of the given camera
func (c *Orthographic) Aspect() float32 {
	return c.right / c.top
}

// SetAspect sets the aspect of the given camera
func (c *Orthographic) SetAspect(width, height float32) {
	c.top = height / 2
	c.bottom = -height / 2
	c.right = width / 2
	c.left = -width / 2
}

// Position returns the position of the given camera.
func (c *Orthographic) Position() math.Vec3 {
	return c.position
}

// SetPosition sets the position of the given camera.
func (c *Orthographic) SetPosition(p math.Vec3) {
	c.position = p
}

// LookAt returns the look at target and up direction of the given camera.
func (c *Orthographic) LookAt() (target, up math.Vec3) {
	target = c.target
	up = c.up
	return
}

// SetLookAt sets the position of the given camera.
func (c *Orthographic) SetLookAt(target, up math.Vec3) {
	c.target = target
	c.up = up
}

// ViewMatrix returns the view matrix of the given camera. The view
// matrix transforms and places the camera up to +Y and towards -Z axis
// at origin.
func (c *Orthographic) ViewMatrix() math.Mat4 {
	l := c.target.Sub(c.position).Unit()
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
