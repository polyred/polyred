// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import "image"

type Option func(*config)

// Title sets the title of the window.
func Title(t string) Option {
	return func(cfg *config) {
		cfg.title = t
	}
}

// MaxSize sets the maximum size of the window.
func MaxSize(w, h int) Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(cfg *config) {
		cfg.maxSize = image.Point{X: w, Y: h}
	}
}

// MinSize sets the minimum size of the window.
func MinSize(w, h int) Option {
	if w <= 0 {
		panic("width must be larger than or equal to 0")
	}
	if h <= 0 {
		panic("height must be larger than or equal to 0")
	}
	return func(cfg *config) {
		cfg.minSize = image.Point{X: w, Y: h}
	}
}

// FPS enables FPS indicator on the top left corner of the window.
func FPS(enable bool) Option {
	return func(cfg *config) {
		cfg.fps = enable
	}
}
