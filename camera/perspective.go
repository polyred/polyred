// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import (
	"poly.red/math"
	"poly.red/object"
)

// Perspective prepresents a perspective camera.
type Perspective struct {
	math.TransformContext

	position math.Vec3
	target   math.Vec3
	up       math.Vec3
	aspect   float64
	fov      float64
	near     float64 // 0 < near < far
	far      float64
}

// NewPerspective creates a new perspective camera with the provided
// camera parameters.
func NewPerspective(opts ...Opt) Interface {
	c := &Perspective{
		position: math.NewVec3(0, 0, 1),
		target:   math.NewVec3(0, 0, 0),
		up:       math.NewVec3(0, 1, 0),
		aspect:   16.0 / 9,
		fov:      60,
		near:     0.01, far: 1000,
	}
	for _, opt := range opts {
		opt(c)
	}

	c.ResetContext()
	return c
}

// Type returns the object type, i.e. object.TypeCamera
func (c *Perspective) Type() object.Type {
	return object.TypeCamera
}

// Fov returns the field of view of the given camera
func (c *Perspective) Fov() float64 {
	return c.fov
}

// Aspect returns the aspect of the given camera
func (c *Perspective) Aspect() float64 {
	return c.aspect
}

// SetAspect sets the aspect of the given camera
func (c *Perspective) SetAspect(width, height float64) {
	c.aspect = width / height
}

// Position returns the position of the given camera.
func (c *Perspective) Position() math.Vec3 {
	return c.position
}

// SetPosition sets the position of the given camera.
func (c *Perspective) SetPosition(p math.Vec3) {
	c.position = p
}

// LookAt returns the look at target and up direction of the given camera.
func (c *Perspective) LookAt() (target, up math.Vec3) {
	target = c.target
	up = c.up
	return
}

// SetLookAt sets the position of the given camera.
func (c *Perspective) SetLookAt(target, up math.Vec3) {
	c.target = target
	c.up = up
}

// ViewMatrix returns the view matrix of the given camera. The view
// matrix transforms and places the camera up to +Y and towards -Z axis
// at origin.
func (c *Perspective) ViewMatrix() math.Mat4 {
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
// After applying projection matrix, the z values are sitting in range
// of [-1, 1] where 1 is the near plane and -1 is the far plane.
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
