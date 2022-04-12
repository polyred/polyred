// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"image"
	"image/color"

	"poly.red/internal/font"
	"poly.red/math"
)

// Window is a minimum abstraction of a Window.
type Window any

// SizeHandler is an extended interface of a Window
// which reports the desired size of a window.
type SizeHandler interface {
	Window
	Size() (w, h int)
}

// DrawHandler is an extended interface of a Window
// which presents a method to draw an image to show
// on top of the window.
type DrawHandler interface {
	Window
	Draw() (screen *image.RGBA, needRedraw bool)
}

// ResizeHandler is an extended Interface of a Window
// which presents a method to handle window resizes.
type ResizeHandler interface {
	Window
	OnResize(w, h int)
}

// KeyboardHanlder is an extended Interface of a Window
// which presents a method to handle keyboard inputs.
type KeyboardHanlder interface {
	Window
	OnKey(key KeyEvent)
}

// MouseHandler is an extended Interface of a Window
// which presents a method to handle mouse inputs.
type MouseHandler interface {
	Window
	OnMouse(mo MouseEvent)
}

// Run runs a object that implements Window interface.
// The window can be configured by a list of options.
func Run(instance Window, opts ...Option) {
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
	width, height := 800, 600
	if ins, ok := instance.(SizeHandler); ok {
		width, height = ins.Size()
	}

	go w.run(instance, config{
		title:   "polyred",
		size:    image.Pt(width, height),
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
