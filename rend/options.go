// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package rend

import (
	"image"
	"image/color"
	"sync"

	"changkun.de/x/ddd/utils"
)

type Option func(opts *Renderer)

func WithSize(width, height int) Option {
	return func(r *Renderer) {
		r.width = width
		r.height = height
	}
}

func WithScene(s *Scene) Option {
	return func(r *Renderer) {
		r.scene = s
	}
}

func WithBackground(c color.RGBA) Option {
	return func(r *Renderer) {
		r.background = c
	}
}

func WithMSAA(n int) Option {
	return func(r *Renderer) {
		r.msaa = n
	}
}

func WithShadowMap(enable bool) Option {
	return func(r *Renderer) {
		r.useShadowMap = enable
	}
}

func WithDebug(enable bool) Option {
	return func(r *Renderer) {
		r.debug = enable
	}
}

func WithConcurrency(n int32) Option {
	return func(r *Renderer) {
		r.concurrentSize = n
	}
}

func WithThreadLimit(n int) Option {
	return func(r *Renderer) {
		r.gomaxprocs = n
	}
}

func (r *Renderer) UpdateOptions(opts ...Option) {
	r.wait() // wait last frame to finish

	for _, opt := range opts {
		opt(r)
	}

	// calibrate rendering size
	r.width *= r.msaa
	r.height *= r.msaa
	r.lockBuf = make([]sync.Mutex, r.width*r.height)
	r.gBuf = make([]gInfo, r.width*r.height)
	r.frameBuf = image.NewRGBA(image.Rect(0, 0, r.width, r.height))
	r.limiter = utils.NewLimiter(r.gomaxprocs)
	r.resetGBuf()
	r.resetFrameBuf()
}
