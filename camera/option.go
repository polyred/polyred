// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package camera

import "poly.red/math"

// Opt represents camera options.
type Opt func(Interface)

// Position sets the camera position.
func Position(pos math.Vec3[float32]) Opt {
	return func(i Interface) {
		i.SetPosition(pos)
	}
}

// LookAt sets the camera look at target and up direction.
func LookAt(target, up math.Vec3[float32]) Opt {
	return func(i Interface) {
		i.SetLookAt(target, up)
	}
}

// ViewFrustum sets the perspective related camera parameters.
//
// If the frustum is using for a perspective camera, the parameters
// must be supply in this order: fov, aspect, near, far
//
// If the frustum is using for a orthographic camera, the parameters
// must be suuply in this order: left, right, bottom, top, near, far
func ViewFrustum(params ...float32) Opt {
	return func(i Interface) {
		switch ii := i.(type) {
		case *Perspective:
			if len(params) != 4 {
				panic("camera: invalid parameter list, expect 4 for perspective camera")
			}
			ii.fov = params[0]
			ii.aspect = params[1]
			ii.near = params[2]
			ii.far = params[3]
		case *Orthographic:
			if len(params) != 6 {
				panic("camera: invalid parameter list, expect 6 for orthographic camera")
			}
			ii.left = params[0]
			ii.right = params[1]
			ii.bottom = params[2]
			ii.top = params[3]
			ii.near = params[4]
			ii.far = params[5]
		default:
			panic("camera: invalid type of camera")
		}
	}
}
