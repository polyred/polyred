// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

import "math"

// CameraType represents the camera type
type CameraType int

const (
	// Orthorgraphic camera type
	Orthorgraphic CameraType = iota
	// Perspective camera type
	Perspective
)

// Camera is a camera interface
type Camera interface {
	Type() CameraType
	GetPosition() Vector
	GetLookAt() Vector
	GetUp() Vector
	GetProjectionMatrix() Matrix
}

// OrthorgraphicCamera represents an orthorgraphic camera.
type OrthorgraphicCamera struct {
	Position Vector
	LookAt   Vector
	Up       Vector
	Left     float64
	Right    float64
	Top      float64
	Bottom   float64
	Near     float64
	Far      float64
}

// Type returns camera type of a given camera
func (c OrthorgraphicCamera) Type() CameraType {
	return Orthorgraphic
}

// GetPosition returns the camera position
func (c OrthorgraphicCamera) GetPosition() Vector {
	return c.Position
}

// GetLookAt returns the camera look at position
func (c OrthorgraphicCamera) GetLookAt() Vector {
	return c.LookAt
}

// GetUp returns the camera up direction
func (c OrthorgraphicCamera) GetUp() Vector {
	return c.Up
}

// GetProjectionMatrix returns the projection matrix
func (c OrthorgraphicCamera) GetProjectionMatrix() Matrix {
	m := NewMatrix()
	m.Set(
		2/(c.Right-c.Left), 0, 0, (c.Left+c.Right)/(c.Left-c.Right),
		0, 2/(c.Top-c.Bottom), 0, (c.Bottom+c.Top)/(c.Bottom-c.Top),
		0, 0, 2/(c.Near-c.Far), (c.Far+c.Near)/(c.Far-c.Near),
		0, 0, 0, 1,
	)
	return m
}

// NewOrthographicCamera returns a new orthographic camera with the given params
func NewOrthographicCamera(Position, LookAt, Up Vector, Left, Right, Top, Bottom, Near, Far float64) Camera {
	return &OrthorgraphicCamera{
		Position, LookAt, Up, Left, Right, Top, Bottom, Near, Far,
	}
}

// PerspectiveCamera implements a perspective camera
type PerspectiveCamera struct {
	Position Vector
	LookAt   Vector
	Up       Vector
	Aspect   float64
	Near     float64
	Far      float64
	FOV      float64
}

// NewPerspectiveCamera returns a new perspective camera with the given params
func NewPerspectiveCamera(Position, LookAt, Up Vector, Aspect, Near, Far, FOV float64) Camera {
	return &PerspectiveCamera{
		Position: Position,
		LookAt:   LookAt,
		Up:       Up,
		Aspect:   Aspect,
		Near:     Near,
		Far:      Far,
		FOV:      FOV,
	}
}

// Type returns camera type of a given camera
func (c PerspectiveCamera) Type() CameraType {
	return Perspective
}

// GetPosition returns the camera position
func (c PerspectiveCamera) GetPosition() Vector {
	return c.Position
}

// GetLookAt returns the camera look at position
func (c PerspectiveCamera) GetLookAt() Vector {
	return c.LookAt
}

// GetUp returns the camera up direction
func (c PerspectiveCamera) GetUp() Vector {
	return c.Up
}

// GetProjectionMatrix returns the projection matrix
func (c PerspectiveCamera) GetProjectionMatrix() Matrix {
	m := NewMatrix()
	m.Set(
		-1/(c.Aspect*math.Tan(c.FOV*math.Pi/360)), 0, 0, 0,
		0, -1/(math.Tan(c.FOV*math.Pi/360)), 0, 0,
		0, 0, (c.Near+c.Far)/(c.Near-c.Far),
		2*(c.Near*c.Far)/(c.Near-c.Far),
		0, 0, 1, 0,
	)
	return m
}
