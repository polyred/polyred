// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package gui implements a minimum windowing utility that collaborate
// with render.Buffer.
//
// A basic window program is as follows:
//
// 	package main
//
// 	import (
// 		"image"
//
// 		"poly.red/gui"
// 		"poly.red/render"
// 	)
//
// 	func main() {
// 		gui.InitWindow()
// 		gui.MainLoop(func(buf *render.Buffer) *image.RGBA {
// 			return nil
// 		})
// 	}
//
package gui

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"poly.red/render"
)

func init() {
	runtime.LockOSThread()
}

type win struct {
	win           *glfw.Window
	title         string
	width, height uint32
	scaleX        float64
	scaleY        float64

	// buffers and draw queue
	buflen int
	bufs   []*render.Buffer
	draw   chan *image.RGBA
	drawQ  chan *image.RGBA

	// Settings
	showFPS bool
	drawer  *font.Drawer

	// Events
	dispatcher *dispatcher
	evSize     SizeEvent
	evMouse    MouseEvent
	evCursor   CursorEvent
	evScroll   ScrollEvent
	evKey      KeyEvent
	mods       ModifierKey

	driverInfo
}

var window *win

// Window returns the window instance
func Window() *win {
	if window != nil {
		return window
	}
	panic("must call gui.InitWindow() first")
}

func Show(img *image.RGBA) {
	if err := InitWindow(); err != nil {
		panic(err)
	}

	MainLoop(func(buf *render.Buffer) *image.RGBA {
		return img
	})
}

// InitWindow constructs a new graphical window.
func InitWindow(opts ...Option) error {
	w := &win{
		title:      "polyred-gui",
		width:      500,
		height:     500,
		showFPS:    false,
		dispatcher: newDispatcher(),
		buflen:     2, // use two buffers by default.
	}
	for _, opt := range opts {
		opt(w)
	}

	w.drawer = &font.Drawer{
		Dst:  nil,
		Src:  image.NewUniform(color.RGBA{200, 100, 0, 255}),
		Face: basicfont.Face7x13,
		Dot:  fixed.P(0*64, 13*64),
	}

	err := glfw.Init()
	if err != nil {
		return fmt.Errorf("failed to initialize glfw context: %w", err)
	}
	w.initWinHints()

	w.win, err = glfw.CreateWindow(int(w.width), int(w.height), w.title, nil, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create glfw window: %w", err))
	}

	// Ratio test. for high DPI, e.g. macOS Retina
	fbw, fbh := w.win.GetFramebufferSize()
	w.scaleX = float64(fbw) / float64(w.width)
	w.scaleY = float64(fbh) / float64(w.height)

	w.draw = make(chan *image.RGBA)
	w.drawQ = make(chan *image.RGBA)
	w.bufs = make([]*render.Buffer, w.buflen)
	for i := 0; i < w.buflen; i++ {
		w.bufs[i] = render.NewBuffer(image.Rect(0, 0, fbw, fbh))
	}

	if window != nil {
		panic("gui: double window initialization")
	}

	window = w
	return nil
}

// Subscribe subscribes the given event and registers the given callback.
func (w *win) Subscribe(eventName EventName, cb EventCallBack) {
	w.dispatcher.eventMap[eventName] = append(w.dispatcher.eventMap[eventName], subscription{
		id: nil,
		cb: cb,
	})
}

// MainLoop executes the given f on a loop, then schedules the returned
// image and renders it to the created window.
func MainLoop(f func(buf *render.Buffer) *image.RGBA) {
	w := window

	// Rendering Thread
	go func() {
		for !w.win.ShouldClose() {
			// We use multiple switching buffers for the drawing, which
			// similar to the double- tripple-buffering techniques.
			// The benefit is that this enables motion vectors between
			// frames.
			//
			// TODO: while executing the rendering on buf2, the buf1
			// is not cleared yet. It should be safe for accessing
			// as previous frame, in order to compute motion vectors.
			// Figuring out what is a proper API design here.
			//
			// A possible design:
			//
			// func MainLoop(f func(buf, prevBuf *render.Buffer) *image.RGBA)
			//
			// Yet there are no enough practice regards the drawbacks
			// of the API, implement a motion vector related algorithm
			// might worthy. e.g. TAA??
			//
			// Maybe we can make framebuf abstraction to be a linked list,
			// in this way, the API remains one buffer parameter, but
			// be able to access previous frames using frame.Prev()?
			for i := range w.bufs {
				w.bufs[i].Clear()
				w.drawQ <- f(w.bufs[i])
			}
		}
	}()

	// Auxiliary Rendering Thread
	//
	// This thread processes the auxiliary informations, such as fps, etc.
	go func() {
		last := time.Now()
		tPerFrame := time.Second / 240 // permit maximum 120 fps
		for buf := range w.drawQ {
			c := time.Now()
			t := c.Sub(last)
			if t < tPerFrame || buf == nil {
				continue
			}
			if w.showFPS {
				w.drawer.Dot = fixed.P(5, 15)
				w.drawer.Dst = buf
				w.drawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}
			last = c
			w.draw <- buf
		}
	}()

	// The Main Thread
	//
	// The main (event) thread terminates when the window instance is
	// closed. All events are handled in the ticked loop.
	//
	// The ticker ticks every ~1ms which permits a maximum of 960 fps
	// (should large enough) for input events handling as the key to
	// making sure the window being responsive (especially on macOS).
	// Since we manage time event timeout ourselves using the ticker,
	// the glfw.PollEvents is used.
	w.initDriver()
	w.initCallbacks()

	ti := time.NewTicker(time.Second / 960)
	for !w.win.ShouldClose() {
		select {
		case buf := <-w.draw:
			// flush is a platform dependent function which uses
			// different drivers. For instance, it use Metal on darwin,
			// OpenGL for Linux and Windows (for now).
			w.flush(buf)
		case <-ti.C:
			glfw.PollEvents()
		}
	}

	w.win.Destroy()
	glfw.Terminate()
}
