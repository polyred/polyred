// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import "poly.red/camera"

// Map is a hard shadow map.
type Map struct {
	camera camera.Interface
	bias   float32
}

// Opt represents a shadow option
type Opt func(sm *Map)

// Camera specifies a camera for rendering a shadow map.
func Camera(c camera.Interface) Opt {
	return func(sm *Map) {
		sm.camera = c
	}
}

// Bias specifies a shadow bias.
func Bias(bias float32) Opt {
	return func(sm *Map) {
		sm.bias = bias
	}
}

// NewMap creates a new shadow map.
func NewMap(opts ...Opt) *Map {
	sm := &Map{
		camera: nil, // default left nil to allow rasterizer decide at runtime
		bias:   0.03,
	}
	for _, opt := range opts {
		opt(sm)
	}
	return sm
}

// Camera returns the camera being used for rendering shadows.
func (sm *Map) Camera() camera.Interface {
	return sm.camera
}

// Bias returns the current shadow bias.
func (sm *Map) Bias() float32 {
	return sm.bias
}
