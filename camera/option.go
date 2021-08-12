// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import "poly.red/math"

// Opt represents camera options.
type Opt func(Interface)

// Position sets the camera position.
func Position(pos math.Vec3) Opt {
	return func(i Interface) {
		i.SetPosition(pos)
	}
}

// LookAt sets the camera look at target and up direction.
func LookAt(target, up math.Vec3) Opt {
	return func(i Interface) {
		i.SetLookAt(target, up)
	}
}

// PerspFrustum sets the perspective related camera parameters.
func PerspFrustum(fov, aspect, near, far float64) Opt {
	return func(i Interface) {
		switch ii := i.(type) {
		case *Perspective:
			ii.fov = fov
			ii.aspect = aspect
			ii.near = near
			ii.far = far
		default:
			panic("camera: misuse of the init option")
		}
	}
}

// OrthoFrustum sets the perspective related camera parameters.
func OrthoFrustum(left, right, bottom, top, near, far float64) Opt {
	return func(i Interface) {
		switch ii := i.(type) {
		case *Orthographic:
			ii.left = left
			ii.right = right
			ii.bottom = bottom
			ii.top = top
			ii.near = near
			ii.far = far
		default:
			panic("camera: misuse of the init option")
		}
	}
}
