// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package app provides the ability to run a window, and handle input
// events.
package app

import (
	"image"
	"image/color"

	"poly.red/internal/font"
	"poly.red/math"
)

// Window is a minimum abstraction of a Window.
type Window interface {
	Size() (w, h int)
}

// DrawHandler is an extended interface of a Window
// which presents a method to draw an image to show
// on top of the window.
type DrawHandler interface {
	Window
	Draw() (screen *image.RGBA, reDraw bool)
}

// ResizeHandler is an extended Interface of a Window
// which presents a method to handle window resizes.
type ResizeHandler interface {
	Window
	OnResize(w, h int)
}

// KeyboardHalder is an extended Interface of a Window
// which presents a method to handle keyboard inputs.
type KeyboardHalder interface {
	Window
	OnKey(key KeyEvent)
}

// KeyboardHalder is an extended Interface of a Window
// which presents a method to handle mouse inputs.
type MouseHandler interface {
	Window
	OnMouse(mo MouseEvent)
}

func Run(instance Window, opts ...Opt) {
	w := &window{
		ready:    make(chan event),
		resize:   make(chan resizeEvent),
		mouse:    make(chan MouseEvent),
		keyboard: make(chan KeyEvent),
		fontDrawer: &font.Drawer{
			Dst:  nil,
			Src:  image.NewUniform(color.RGBA{200, 100, 0, 255}),
			Face: font.Face7x13,
			Dot:  math.P(0*64, 13*64),
		},
	}
	go w.run(instance, config{
		title:   "polyred",
		size:    image.Pt(instance.Size()),
		minSize: image.Pt(50, 50),
		maxSize: image.Pt(1920*2, 1080*2),
		fps:     false,
	}, opts...)
	w.main(instance)
}

type config struct {
	title   string
	size    image.Point
	maxSize image.Point
	minSize image.Point
	fps     bool
}

type window struct {
	win        *osWindow
	ready      chan event
	keyboard   chan KeyEvent
	mouse      chan MouseEvent
	resize     chan resizeEvent
	fontDrawer *font.Drawer
}

type event struct{}
type resizeEvent struct{ w, h int }