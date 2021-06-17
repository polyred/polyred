// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import "changkun.de/x/ddd/camera"

type ShadowType int

const (
	ShadowTypeHard ShadowType = iota // hard shadow mapping
	ShadowTypePCF                    // percentage closer filtering
	ShadowTypePCSS                   // percentage closer soft shadows
	ShadowTypeVSSM                   // variance soft shadow mapping
	ShadowTypeMSM                    // moment shadow mapping
)

type ShadowMap struct {
	typ    ShadowType
	camera camera.Interface
	bias   float64
}

type ShadowMapOption func(sm *ShadowMap)

func WithShadowMapType(typ ShadowType) ShadowMapOption {
	return func(sm *ShadowMap) {
		sm.typ = typ
	}
}

func WithShadowMapCamera(c camera.Interface) ShadowMapOption {
	return func(sm *ShadowMap) {
		sm.camera = c
	}
}

func WithShadowMapBias(bias float64) ShadowMapOption {
	return func(sm *ShadowMap) {
		sm.bias = bias
	}
}

func NewShadowMap(opts ...ShadowMapOption) *ShadowMap {
	sm := &ShadowMap{
		typ:    ShadowTypeHard,
		camera: nil, // default left nil to allow rasterizer decide at runtime
		bias:   0.03,
	}
	for _, opt := range opts {
		opt(sm)
	}
	return sm
}

func (sm *ShadowMap) Camera() camera.Interface {
	return sm.camera
}

func (sm *ShadowMap) Bias() float64 {
	return sm.bias
}
