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
	"poly.red/internal/font"
	"poly.red/math"
	"poly.red/render"
	"poly.red/texture/buffer"
)

func init() {
	runtime.LockOSThread()
}

type Window struct {
	win           *glfw.Window
	title         string
	width, height int
	scaleX        float32
	scaleY        float32

	// renderer and draw queue
	renderer *render.Renderer
	drawQ    chan *buffer.Buffer
	resize   chan image.Rectangle

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
	r := render.NewRenderer(render.Size(img.Bounds().Dx(), img.Bounds().Dy()))
	w, err := NewWindow(r)
	if err != nil {
		panic(err)
	}

	buf := r.NextBuffer()
	r.DrawImage(buf, img)
	w.MainLoop(func() *buffer.Buffer { return buf })
}

// NewWindow constructs a new graphical window.
func NewWindow(r *render.Renderer, opts ...Option) (*Window, error) {
	w := &Window{
		title:      "polyred-gui",
		width:      r.CurrBuffer().Bounds().Dx(),
		height:     r.CurrBuffer().Bounds().Dy(),
		showFPS:    false,
		dispatcher: newDispatcher(),
		renderer:   r,
	}
	for _, opt := range opts {
		opt(w)
	}

	w.drawer = &font.Drawer{
		Dst:  nil,
		Src:  image.NewUniform(color.RGBA{200, 100, 0, 255}),
		Face: font.Face7x13,
		Dot:  math.P(0*64, 13*64),
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
	w.scaleX = float32(fbw) / float32(w.width)
	w.scaleY = float32(fbh) / float32(w.height)

	w.drawQ = make(chan *buffer.Buffer)
	w.resize = make(chan image.Rectangle)
	r.Options(render.Size(fbw, fbh))
	return w, nil
}

// Subscribe subscribes the given event and registers the given callback.
func (w *Window) Subscribe(eventName EventName, cb EventCallBack) {
	w.dispatcher.eventMap[eventName] = append(w.dispatcher.eventMap[eventName], subscription{
		id: nil,
		cb: cb,
	})
}

type frameBuf struct {
	img  *image.RGBA
	done chan struct{}
}

// MainLoop executes the given f on a loop, then schedules the returned
// image and renders it to the created window.
func (w *Window) MainLoop(f func() *buffer.Buffer) {
	// Rendering Thread
	go func() {
		for !w.win.ShouldClose() {
			select {
			case r := <-w.resize:
				w.renderer.Options(render.Size(w.width, w.height))
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
				w.drawQ <- f()
			}
		}
	}()

	// Auxiliary Rendering Thread
	//
	// This thread processes the auxiliary informations, such as fps, etc.
	go func() {
		runtime.LockOSThread()
		w.initContext()

		// Triple buffering
		bufs := [3]*frameBuf{{}, {}, {}}
		bufIdx := 0

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
				w.drawer.Dot = math.P(5, 15)
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
			fbuf := bufs[bufIdx%3]
			fbuf.img = buf.Image()
			w.flush(fbuf)
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
		w.width = width
		w.height = height
		w.scaleX = float32(fbw) / float32(width)
		w.scaleY = float32(fbh) / float32(height)
		w.dispatcher.Dispatch(OnResize, &w.evSize)
		w.resize <- image.Rect(0, 0, fbw, fbh)
	})
	w.win.SetCursorPosCallback(func(_ *glfw.Window, xpos, ypos float64) {
		w.evCursor.Xpos = float32(xpos)
		w.evCursor.Ypos = float32(ypos)
		w.evCursor.Mods = w.mods
		w.dispatcher.Dispatch(OnCursor, &w.evCursor)
	})
	w.win.SetMouseButtonCallback(func(x *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		xpos, ypos := x.GetCursorPos()
		w.evMouse.Button = MouseButton(button)
		w.evMouse.Mods = ModifierKey(mods)
		w.evMouse.Xpos = float32(xpos)
		w.evMouse.Ypos = float32(ypos)

		switch action {
		case glfw.Press:
			w.dispatcher.Dispatch(OnMouseDown, &w.evMouse)
		case glfw.Release:
			w.dispatcher.Dispatch(OnMouseUp, &w.evMouse)
		}
	})
	w.win.SetScrollCallback(func(_ *glfw.Window, xoff, yoff float64) {
		w.evScroll.Xoffset = float32(xoff)
		w.evScroll.Yoffset = float32(yoff)
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
