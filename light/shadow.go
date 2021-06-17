// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package light

import "changkun.de/x/ddd/camera"

type ShadowType int

const (
	ShadowTypePCF  ShadowType = iota // percentage closer filtering
	ShadowTypePCSS                   // percentage closer soft shadows
	ShadowTypeVSSM                   // variance soft shadow mapping
	ShadowTypeMSM                    // moment shadow mapping
)

type ShadowMap struct {
	typ    ShadowType
	camera camera.Interface
	bias   float64
	size   int
	tex    []float64
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

func WithShadowMapSize(size int) ShadowMapOption {
	return func(sm *ShadowMap) {
		sm.size = size
	}
}

func NewShadowMap(opts ...ShadowMapOption) *ShadowMap {
	sm := &ShadowMap{
		typ:    ShadowTypePCF,
		camera: nil, // default left nil to allow rasterizer decide at runtime
		bias:   0,
		size:   512,
	}
	for _, opt := range opts {
		opt(sm)
	}
	sm.tex = make([]float64, sm.size)
	return sm
}
