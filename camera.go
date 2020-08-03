// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a GNU GPLv3
// license that can be found in the LICENSE file.

package ddd

// PerspectiveCamera implements a perspective camera.
type PerspectiveCamera struct {
	Position Vector
	LookAt   Vector
	Up       Vector
	Aspect   float64
	Near     float64
	Far      float64
	FOV      float64
}

// NewPerspectiveCamera returns a new perspective camera with the given params.
func NewPerspectiveCamera(Position, LookAt, Up Vector, Aspect, Near, Far, FOV float64) *PerspectiveCamera {
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
