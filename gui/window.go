// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"runtime"
	"time"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.design/x/mainthread"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Option is a functional option to the window constructor New.
type Option func(*win)

// WithTitle option sets the title (caption) of the window.
func WithTitle(title string) Option {
	return func(o *win) {
		o.title = title
	}
}

// WithSize option sets the width and height of the window.
func WithSize(width, height int) Option {
	return func(o *win) {
		o.width = width
		o.height = height
	}
}

// WithFPS sets the window to show FPS.
func WithFPS() Option {
	return func(o *win) {
		o.showFPS = true
	}
}

type win struct {
	win           *glfw.Window
	title         string
	width, height int
	ratio         int // for retina display
	showFPS       bool
	drawer        *font.Drawer
	last          time.Time
}

// NewWindow constructs a new graphical window.
func NewWindow(opts ...Option) *win {
	w := &win{
		title:   "",
		width:   500,
		height:  500,
		showFPS: false,
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

	var err error
	defer func() {
		if err != nil {
			// This function must be called from the mainthread.
			mainthread.Call(w.win.Destroy)
			log.Fatalf("failed to create a window: %v", err)
		}
	}()

	mainthread.Call(func() {
		err = glfw.Init()
		if err != nil {
			log.Fatalf("failed to initialize glfw context: %v", err)
		}
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
		glfw.WindowHint(glfw.Resizable, glfw.False)

		w.win, err = glfw.CreateWindow(w.width, w.height, w.title, nil, nil)
		if err != nil {
			return
		}

		// Ratio test. for high DPI, e.g. macOS Retina
		width, _ := w.win.GetFramebufferSize()
		w.ratio = width / w.width
		if w.ratio < 1 {
			w.ratio = 1
		}
		w.win.Destroy()

		w.win, err = glfw.CreateWindow(w.width/w.ratio, w.height/w.ratio, w.title, nil, nil)
	})
	if err != nil {
		return nil
	}
	err = gl.Init()
	if err != nil {
		return nil
	}

	return w
}

func (w *win) MainLoop(f func() *image.RGBA) {
	defer func() {
		// This function must be called from the mainthread.
		mainthread.Call(w.win.Destroy)
	}()

	go func() {
		runtime.LockOSThread()
		w.win.MakeContextCurrent()
		for !w.win.ShouldClose() {
			w.flush(f())
		}
	}()

	for !w.win.ShouldClose() {
		mainthread.Call(func() {
			glfw.WaitEvents()
			// or:
			// glfw.WaitEventsTimeout(1.0 / 30)
		})
	}
}

func (w *win) flush(img *image.RGBA) {
	if img == nil {
		return
	}
	if w.showFPS {
		w.drawer.Dot = fixed.P(5, 15)
		w.drawer.Dst = img
		w.drawer.DrawString(fmt.Sprintf("%d", time.Second/time.Since(w.last)))
		w.last = time.Now()
	}

	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	gl.DrawBuffer(gl.FRONT)
	gl.Viewport(0, 0, int32(dx), int32(dy))
	gl.RasterPos2d(-1, 1)
	gl.PixelZoom(1, -1)
	gl.DrawPixels(int32(dx), int32(dy), gl.RGBA, gl.UNSIGNED_BYTE,
		unsafe.Pointer(&img.Pix[0]))
	gl.Flush()
}
