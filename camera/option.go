package camera

import "poly.red/math"

// Option represents camera options.
type Option func(Interface)

// WithPosition sets the camera position.
func WithPosition(pos math.Vec3) Option {
	return func(i Interface) {
		i.SetPosition(pos)
	}
}

// WithLookAt sets the camera look at target and up direction.
func WithLookAt(target, up math.Vec3) Option {
	return func(i Interface) {
		i.SetLookAt(target, up)
	}
}

// WithPerspFrustum sets the perspective related camera parameters.
func WithPerspFrustum(fov, aspect, near, far float64) Option {
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

// WithOrthoFrustum sets the perspective related camera parameters.
func WithOrthoFrustum(left, right, bottom, top, near, far float64) Option {
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
