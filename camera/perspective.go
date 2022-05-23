// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import (
	"poly.red/geometry/primitive"
	"poly.red/math"
	"poly.red/scene/object"
)

// Perspective prepresents a perspective camera.
type Perspective struct {
	math.TransformContext[float32]

	position math.Vec3[float32]
	target   math.Vec3[float32]
	up       math.Vec3[float32]
	aspect   float32
	fov      float32
	near     float32 // 0 < near < far
	far      float32
}

// NewPerspective creates a new perspective camera with the provided
// camera parameters.
func NewPerspective(opts ...Option) Interface {
	c := &Perspective{
		position: math.NewVec3[float32](0, 0, 1),
		target:   math.NewVec3[float32](0, 0, 0),
		up:       math.NewVec3[float32](0, 1, 0),
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

func (c *Perspective) Name() string { return "perspective_camera" }

// Type returns the object type, i.e. object.TypeCamera
func (c *Perspective) Type() object.Type {
	return object.TypeCamera
}

// Fov returns the field of view of the given camera
func (c *Perspective) Fov() float32 {
	return c.fov
}

// Aspect returns the aspect of the given camera
func (c *Perspective) Aspect() float32 {
	return c.aspect
}

// SetAspect sets the aspect of the given camera
func (c *Perspective) SetAspect(width, height float32) {
	c.aspect = width / height
}

// Position returns the position of the given camera.
func (c *Perspective) Position() math.Vec3[float32] {
	return c.position
}

// SetPosition sets the position of the given camera.
func (c *Perspective) SetPosition(p math.Vec3[float32]) {
	c.position = p
}

// LookAt returns the look at target and up direction of the given camera.
func (c *Perspective) LookAt() (target, up math.Vec3[float32]) {
	target = c.target
	up = c.up
	return
}

// SetLookAt sets the position of the given camera.
func (c *Perspective) SetLookAt(target, up math.Vec3[float32]) {
	c.target = target
	c.up = up
}

// ViewMatrix returns the view matrix of the given camera. The view
// matrix transforms and places the camera up to +Y and towards -Z axis
// at origin.
func (c *Perspective) ViewMatrix() math.Mat4[float32] {
	return ViewMatrix(c.position, c.target, c.up)
}

// ProjMatrix returns the projection matrix of the given camera.
// After applying projection matrix, the z values are sitting in range
// of [-1, 1] where 1 is the near plane and -1 is the far plane.
func (c *Perspective) ProjMatrix() math.Mat4[float32] {
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

func (c *Perspective) AABB() primitive.AABB { return primitive.NewAABB(c.position) }
