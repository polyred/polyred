// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gui

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"changkun.de/x/polyred/render"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

func init() {
	runtime.LockOSThread()
}

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

	bufs  []*render.Buffer
	draw  chan *image.RGBA
	drawQ chan *image.RGBA

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

	err := glfw.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize glfw context: %w", err))
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
	glfw.WindowHint(glfw.Resizable, glfw.True)

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
	w.bufs = make([]*render.Buffer, 2)
	for i := range w.bufs {
		w.bufs[i] = render.NewBuffer(image.Rect(0, 0, fbw, fbh))
	}
	w.win.MakeContextCurrent()
	err = gl.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize gl: %w", err))
	}

	// Setup event callbacks
	w.win.SetSizeCallback(func(x *glfw.Window, width int, height int) {
		fbw, fbh := x.GetFramebufferSize()
		w.evSize.Width = width
		w.evSize.Height = height
		w.scaleX = float64(fbw) / float64(width)
		w.scaleY = float64(fbh) / float64(height)
		w.dispatcher.Dispatch(OnResize, &w.evSize)

		// The following replaces the w.bufs on the main thread.
		//
		// It does not involve with data race. Because the draw call is
		// also handled on the main thread, which is currently not possible
		// to execute.
		r := image.Rect(0, 0, fbw, fbh)
		for i := range w.bufs {
			w.bufs[i] = render.NewBuffer(r)
		}
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

	once.Do(func() { window = w })
}

func (w *win) Subscribe(eventName EventName, cb EventCallBack) {
	w.dispatcher.eventMap[eventName] = append(w.dispatcher.eventMap[eventName], subscription{
		id: nil,
		cb: cb,
	})
}

// MainLoop executes the given f on a loop, then schedules the returned
// image and renders it to the created window. f
func MainLoop(f func(frame *render.Buffer) *image.RGBA) {
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
			// be able to access previous frames using frame.Prev()
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
			current := time.Now()
			t := current.Sub(last)
			if t < tPerFrame || buf == nil {
				continue
			}
			if w.showFPS {
				w.drawer.Dot = fixed.P(5, 15)
				w.drawer.Dst = buf
				w.drawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}
			last = current
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
	w.win.MakeContextCurrent()
	err := gl.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize gl: %w", err))
	}
	gl.DrawBuffer(gl.FRONT)
	gl.PixelZoom(1, -1)

	ti := time.NewTicker(time.Second / 960)
	for !w.win.ShouldClose() {
		select {
		case buf := <-w.draw:
			w.flush(buf)
		case <-ti.C:
			glfw.PollEvents()
		}
	}

	w.win.Destroy()
}

// flush flushes the containing pixel buffer of the given image to the
// hardware frame buffer for display prupose. The given image is assumed
// to be non-nil pointer.
func (w *win) flush(img *image.RGBA) {
	dx, dy := int32(img.Bounds().Dx()), int32(img.Bounds().Dy())
	gl.RasterPos2d(-1, 1)
	gl.Viewport(0, 0, dx, dy)
	gl.DrawPixels(dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))
	gl.Flush()
}
