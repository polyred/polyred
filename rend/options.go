// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package rend

import "image/color"

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
