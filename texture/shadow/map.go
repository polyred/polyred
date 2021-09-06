// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package shadow

import "poly.red/camera"

// Type represents a shadow type
type Type int

// All kinds of shadow types
const (
	ShadowTypeHard Type = iota // hard shadow mapping
	ShadowTypePCF              // percentage closer filtering
	ShadowTypePCSS             // percentage closer soft shadows
	ShadowTypeVSSM             // variance soft shadow mapping
	ShadowTypeMSM              // moment shadow mapping
)

// Map is a shadow map.
type Map struct {
	typ    Type
	camera camera.Interface
	bias   float32
}

// Opt represents a shadow option
type Opt func(sm *Map)

// Method specifies a underlying used algorithm for rendering shadows.
func Method(typ Type) Opt {
	return func(sm *Map) {
		sm.typ = typ
	}
}

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
		typ:    ShadowTypeHard,
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
