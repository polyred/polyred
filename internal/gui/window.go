// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

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
	"poly.red/texture/buffer"
)

func init() {
	runtime.LockOSThread()
}

type Window struct {
	win           *glfw.Window
	title         string
	width, height uint32
	scaleX        float64
	scaleY        float64

	// buffers and draw queue
	buflen int
	bufs   []*buffer.Buffer
	drawQ  chan *image.RGBA
	resize chan image.Rectangle

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

// Show shows the given image on a window.
func Show(img *image.RGBA) {
	// TODO: fix on darwin about scaling issue, compute dynamically
	opt := WithSize(img.Bounds().Dx(), img.Bounds().Dy())
	if runtime.GOOS == "darwin" {
		opt = WithSize(img.Bounds().Dx()/2, img.Bounds().Dy()/2)
	}
	w, err := NewWindow(opt)
	if err != nil {
		panic(err)
	}

	// TODO: rethink about the main loop and speedup. if
	// the callback is returning the same image, why not caching
	// the result?
	//
	// Possible issues: double buffering.
	w.MainLoop(func(buf *buffer.Buffer) *image.RGBA { return img })
}

// NewWindow constructs a new graphical window.
func NewWindow(opts ...Option) (*Window, error) {
	w := &Window{
		title:      "polyred-gui",
		width:      800,
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
		return nil, fmt.Errorf("failed to initialize glfw context: %w", err)
	}
	w.initWinHints()

	w.win, err = glfw.CreateWindow(int(w.width), int(w.height), w.title, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create glfw window: %w", err)
	}

	// Ratio test. for high DPI, e.g. macOS Retina
	fbw, fbh := w.win.GetFramebufferSize()
	w.scaleX = float64(fbw) / float64(w.width)
	w.scaleY = float64(fbh) / float64(w.height)

	w.drawQ = make(chan *image.RGBA)
	w.resize = make(chan image.Rectangle)
	w.bufs = make([]*buffer.Buffer, w.buflen)
	w.resetBufs(image.Rect(0, 0, fbw, fbh))
	return w, nil
}

// Subscribe subscribes the given event and registers the given callback.
func (w *Window) Subscribe(eventName EventName, cb EventCallBack) {
	w.dispatcher.eventMap[eventName] = append(w.dispatcher.eventMap[eventName], subscription{
		id: nil,
		cb: cb,
	})
}

// MainLoop executes the given f on a loop, then schedules the returned
// image and renders it to the created window.
func (w *Window) MainLoop(f func(buf *buffer.Buffer) *image.RGBA) {
	// Rendering Thread
	go func() {
		for !w.win.ShouldClose() {
			select {
			case r := <-w.resize:
				w.resetBufs(r)
			default:
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
				// func MainLoop(f func(buf, prevBuf *render.Buffer) *image.RGBA
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
		}
	}()

	// Auxiliary Rendering Thread
	//
	// This thread processes the auxiliary informations, such as fps, etc.
	go func() {
		runtime.LockOSThread()
		w.initContext()

		last := time.Now()
		tPerFrame := time.Second / 240 // permit maximum 120 fps
		for buf := range w.drawQ {
			c := time.Now()
			t := c.Sub(last)
			if t < tPerFrame || buf == nil {
				continue
			}
			if w.showFPS {
				// FIXME: should draw based on buffer.Buffer format.
				w.drawer.Dot = fixed.P(5, 15)
				w.drawer.Dst = buf
				w.drawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}
			last = c
			// flush is a platform dependent function which uses
			// different drivers. For instance, it use Metal on darwin,
			// OpenGL for Linux and Windows (for now).
			//
			// If the flush is executed on the mainthread, it takes too
			// long to start execute event processing, which may laggy.
			// Hence, we should run it into a non-mainthread call. See:
			// https://golang.design/research/ultimate-channel/
			w.flush(buf)
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
		<-ti.C
		glfw.PollEvents()
	}

	w.win.Destroy()
	glfw.Terminate()
}

func (w *Window) initCallbacks() {
	// Setup event callbacks
	w.win.SetSizeCallback(func(x *glfw.Window, width int, height int) {
		fbw, fbh := x.GetFramebufferSize()
		w.evSize.Width = width
		w.evSize.Height = height
		w.scaleX = float64(fbw) / float64(width)
		w.scaleY = float64(fbh) / float64(height)
		w.dispatcher.Dispatch(OnResize, &w.evSize)
		w.resize <- image.Rect(0, 0, fbw, fbh)
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
}
