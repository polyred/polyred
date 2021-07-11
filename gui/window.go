// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"sync"
	"time"
	"unsafe"

	"changkun.de/x/polyred/render"
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
		o.width = uint32(width)
		o.height = uint32(height)
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
	width, height uint32
	scaleX        float64
	scaleY        float64

	buf1 *render.Buffer
	buf2 *render.Buffer
	draw chan *image.RGBA

	// Settings
	showFPS bool
	drawer  *font.Drawer
	last    time.Time

	// Events
	dispatcher *dispatcher
	evSize     SizeEvent
	evMouse    MouseEvent
	evCursor   CursorEvent
	evScroll   ScrollEvent
	evKey      KeyEvent
	mods       ModifierKey
}

var (
	once   sync.Once
	window *win
)

// Window returns the window instance
func Window() *win {
	if window != nil {
		return window
	}
	panic("must call gui.InitWindow() first")
}

// InitWindow constructs a new graphical window.
func InitWindow(opts ...Option) {
	w := &win{
		title:      "",
		width:      500,
		height:     500,
		showFPS:    false,
		dispatcher: newDispatcher(),
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
			log.Fatalf("window: %v", err)
		}
	}()

	mainthread.Call(func() {
		err = glfw.Init()
		if err != nil {
			err = fmt.Errorf("failed to initialize glfw context: %w", err)
			return
		}
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
		glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
		glfw.WindowHint(glfw.Resizable, glfw.True)

		w.win, err = glfw.CreateWindow(int(w.width), int(w.height), w.title, nil, nil)
		if err != nil {
			err = fmt.Errorf("failed to create glfw window: %w", err)
			return
		}

		// Ratio test. for high DPI, e.g. macOS Retina
		fbw, fbh := w.win.GetFramebufferSize()
		w.scaleX = float64(fbw) / float64(w.width)
		w.scaleY = float64(fbh) / float64(w.height)
		w.buf1 = render.NewBuffer(image.Rect(0, 0, fbw, fbh))
		w.buf2 = render.NewBuffer(image.Rect(0, 0, fbw, fbh))
		w.draw = make(chan *image.RGBA)

		// Make sure this happens on main thread. Otherwise, Windows
		// cannot render anything from it.
		w.win.MakeContextCurrent()
		err = gl.Init()
		if err != nil {
			err = fmt.Errorf("failed to initialize gl: %w", err)
			return
		}
	})
	if err != nil {
		return
	}

	once.Do(func() { window = w })
}

func (w *win) Subscribe(eventName EventName, cb EventCallBack) {
	w.dispatcher.eventMap[eventName] = append(w.dispatcher.eventMap[eventName], subscription{
		id: nil,
		cb: cb,
	})
}

func MainLoop(f func(buf *render.Buffer) *image.RGBA) {
	w := window

	defer func() {
		// This function must be called from the mainthread.
		mainthread.Call(w.win.Destroy)
	}()

	// Setup event callbacks
	w.win.SetSizeCallback(func(x *glfw.Window, width int, height int) {
		fbw, fbh := x.GetFramebufferSize()
		w.evSize.Width = width
		w.evSize.Height = height
		w.scaleX = float64(fbw) / float64(width)
		w.scaleY = float64(fbh) / float64(height)
		w.dispatcher.Dispatch(OnResize, &w.evSize)
		w.refreshBuf(image.Rect(0, 0, fbw, fbh))
	})
	w.win.SetCursorPosCallback(func(_ *glfw.Window, xpos, ypos float64) {
		w.evCursor.Xpos = xpos
		w.evCursor.Ypos = ypos
		w.evCursor.Mods = w.mods
		w.dispatcher.Dispatch(OnCursor, &w.evCursor)
	})
	w.win.SetMouseButtonCallback(func(x *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		xpos, ypos := x.GetCursorPos()
		w.evMouse.Button = MouseButton(button)
		w.evMouse.Mods = ModifierKey(mods)
		w.evMouse.Xpos = xpos
		w.evMouse.Ypos = ypos

		switch action {
		case glfw.Press:
			w.dispatcher.Dispatch(OnMouseDown, &w.evMouse)
		case glfw.Release:
			w.dispatcher.Dispatch(OnMouseUp, &w.evMouse)
		}
	})
	w.win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		w.evScroll.Xoffset = xoff
		w.evScroll.Yoffset = yoff
		w.evScroll.Mods = w.mods
		w.dispatcher.Dispatch(OnScroll, &w.evScroll)
	})
	w.win.SetKeyCallback(func(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		w.evKey.Key = Key(key)
		w.evKey.Mods = ModifierKey(mods)
		w.mods = w.evKey.Mods
		switch action {
		case glfw.Press:
			w.dispatcher.Dispatch(OnKeyDown, &w.evKey)
		case glfw.Release:
			w.dispatcher.Dispatch(OnKeyUp, &w.evKey)
		case glfw.Repeat:
			w.dispatcher.Dispatch(OnKeyRepeat, &w.evKey)
		}
	})

	// Rendering Thread
	go func() {
		// We use two switching buffers for the draw calls. Otherwise,
		// The gl.DrawPixels will race with pixel memories.
		//
		// Consider the following diagram:
		//
		//           +------ buf.Clear()
		//           v
		// |f(w.buf)| |f(w.buf)|
		// +--------+-+-------------------------------------------> Render
		//           \ w.draw <- buf
		//            v
		// -----------+--+----------------------------------------> Event
		//                \ gl.DrawPixels + gl.Flush
		//                 v
		// ----------------+---+----------------------------------> GPU
		//                      \
		//                       v
		// ----------------------+--------------------------------> Monitor
		//          |<-- ~5ms -->|
		//
		// According to a rough measurement, when f finishes the rendering,
		// the time to be able to flush the entire pixel buffer to the
		// monitor requires 5ms on a MacBook Air (M1, 2020) laptop.
		// This means, if the next f(w.buf) is called before flushing
		// the pixels onto monitor, the buf.Clear inside the f happens
		// concurrently with the flushing, aka data race.
		//
		// Although the race is not serious enough since it is just pixels,
		// but it still can cause multiple known issues:
		//
		// 1. Crash on specific platform when resizing the window, such as macOS
		// 2. Black flicking on the rendering even without window resizing
		//
		// To prevent that happen, we use two buffers that call on different
		// draw calls. The benefits, of course, resolve the above issues,
		// and be able to compute motion vectors between frames.
		for !w.win.ShouldClose() {
			w.buf1.Clear()
			w.draw <- f(w.buf1)
			w.buf2.Clear()
			w.draw <- f(w.buf2)
		}
	}()

	// Event Thread
	timeout := 1.0 / 120 // maximum 120 fps
	mainthread.Call(func() {
		gl.DrawBuffer(gl.FRONT)
		gl.PixelZoom(1, -1)
	})
	for !w.win.ShouldClose() {
		mainthread.Call(func() {
			w.flush(<-w.draw)
			glfw.WaitEventsTimeout(timeout)
		})
	}
}

func (w *win) refreshBuf(r image.Rectangle) {
	w.buf1 = render.NewBuffer(r)
	w.buf2 = render.NewBuffer(r)
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

	gl.RasterPos2d(-1, 1)
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	gl.Viewport(0, 0, int32(dx), int32(dy))
	gl.DrawPixels(int32(dx), int32(dy), gl.RGBA,
		gl.UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))
	gl.Flush()
}
