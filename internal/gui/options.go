// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

// Option is a functional option to the window constructor New.
type Option func(*Window)

// WithTitle option sets the title (caption) of the window.
func WithTitle(title string) Option {
	return func(o *Window) {
		o.title = title
	}
}

// WithSize option sets the width and height of the window.
func WithSize(width, height int) Option {
	return func(o *Window) {
		o.width = uint32(width)
		o.height = uint32(height)
	}
}

// WithNumBufs option sets the number of switching buffers are used
// internally for the rendering.
func WithNumBufs(num int) Option {
	return func(w *Window) {
		w.buflen = num
	}
}

// WithFPS sets the window to show FPS.
func WithFPS() Option {
	return func(o *Window) {
		o.showFPS = true
	}
}
