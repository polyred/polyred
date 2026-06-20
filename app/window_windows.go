// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"fmt"
	"image"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"poly.red/gpu/gl"
	"poly.red/gpu/syscall/windows"
	"poly.red/math"
)

type osWindow struct {
	hwnd syscall.Handle
	hdc  syscall.Handle
	ctx  *winContext

	viewScale int
	config    *config
}

var resources struct {
	// handle is the module handle from GetModuleHandle.
	handle syscall.Handle
	// class is the window class from RegisterClassEx.
	class uint16
	// cursor is the arrow cursor resource.
	cursor syscall.Handle
}

// initResources initializes the resources global.
func initResources() error {
	windows.SetProcessDPIAware()
	hInst, err := windows.GetModuleHandle()
	if err != nil {
		return err
	}
	resources.handle = hInst
	c, err := windows.LoadCursor(windows.IDC_ARROW)
	if err != nil {
		return err
	}
	resources.cursor = c

	name, err := syscall.UTF16PtrFromString("polyred")
	if err != nil {
		return err
	}

	wcls := windows.WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(windows.WndClassEx{})),
		Style:         windows.CS_HREDRAW | windows.CS_VREDRAW | windows.CS_OWNDC,
		LpfnWndProc:   syscall.NewCallback(windowProc),
		HInstance:     hInst,
		LpszClassName: name,
	}
	cls, err := windows.RegisterClassEx(&wcls)
	if err != nil {
		return err
	}
	resources.class = cls
	return nil
}

var winMap sync.Map // map[HWND]struct{*window, app.Window}

type winapp struct {
	win *window
	app Window
}

var dead = make(chan struct{})

func (w *window) main(app Window) { <-dead }

func (w *window) run(app Window, cfg config, opts ...Option) {
	// GetMessage and PeekMessage can filter on a window HWND, but
	// then thread-specific messages such as WM_QUIT are ignored.
	// Instead lock the thread so window messages arrive through
	// unfiltered GetMessage calls.
	runtime.LockOSThread()

	initResources()
	viewScale := windows.GetSystemDPI()
	dwStyle := uint32(windows.WS_OVERLAPPEDWINDOW)
	dwExStyle := uint32(windows.WS_EX_APPWINDOW | windows.WS_EX_WINDOWEDGE)

	hwnd, err := windows.CreateWindowEx(dwExStyle,
		resources.class,
		"",
		dwStyle|windows.WS_CLIPSIBLINGS|windows.WS_CLIPCHILDREN,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		0,
		0,
		resources.handle,
		0)
	if err != nil {
		panic(fmt.Errorf("app: failed to create a window: %w", err))
	}
	hdc, err := windows.GetDC(hwnd)
	if err != nil {
		panic(fmt.Errorf("app: failed to create a window: %w", err))
	}

	w.win = &osWindow{
		hwnd:      hwnd,
		hdc:       hdc,
		viewScale: viewScale,
		config:    &cfg,
	}
	winMap.Store(w.win.hwnd, winapp{w, app})
	defer winMap.Delete(w.win.hwnd)
	w.configs(opts...)

	windows.ShowWindow(w.win.hwnd, windows.SW_SHOWDEFAULT)
	windows.SetForegroundWindow(w.win.hwnd)
	windows.SetFocus(w.win.hwnd)

	// The EGL (ANGLE) context is created after the window is shown, mirroring
	// the X11 path which builds the context after the window is mapped.
	w.win.ctx, err = newWinContext(w.win)
	if err != nil {
		panic(fmt.Errorf("egl: cannot create context for window: %w", err))
	}
	if err := w.win.ctx.Refresh(); err != nil {
		panic(fmt.Errorf("egl: cannot create surface: %w", err))
	}

	// A single background goroutine owns rendering; the Win32 message pump runs
	// here on the locked thread. This mirrors the Linux model (go w.draw(app) +
	// event loop) instead of driving draw synchronously from WM_PAINT, which
	// would block the pump.
	go w.draw(app)

	w.event()
	dead <- struct{}{}
}

func (w *window) configs(opts ...Option) {
	cfg := w.win.config
	for _, o := range opts {
		o(cfg)
	}

	width, height := int32(cfg.size.X), int32(cfg.size.Y)

	wr := windows.Rect{
		Right:  int32(width),
		Bottom: int32(height),
	}
	dwStyle := uint32(windows.WS_OVERLAPPEDWINDOW)
	dwExStyle := uint32(windows.WS_EX_APPWINDOW | windows.WS_EX_WINDOWEDGE)
	windows.AdjustWindowRectEx(&wr, dwStyle, 0, dwExStyle)
	windows.MoveWindow(w.win.hwnd, 0, 0, width, height, true)
	windows.SetWindowText(w.win.hwnd, cfg.title)
}

func (w *window) event() {
	msg := new(windows.Msg)
loop:
	for {
		switch ret := windows.GetMessage(msg, 0, 0, 0); ret {
		case -1:
			panic("app: GetMessage failed")
		case 0: // WM_QUIT received.
			break loop
		}
		windows.TranslateMessage(msg)
		windows.DispatchMessage(msg)
	}
}

func (w *window) draw(app Window) {
	// GL calls must stay on a single thread; the EGL context is made current
	// here. Mirrors the X11 draw goroutine.
	runtime.LockOSThread()
	if err := w.win.ctx.Lock(); err != nil {
		panic(fmt.Errorf("egl: cannot make context current: %w", err))
	}
	defer w.win.ctx.Unlock()

	g := w.win.ctx.gl

	vertices := slice2bytes([]float32{
		-1, +1, 0, 0,
		+1, +1, 1, 0,
		-1, -1, 0, 1,
		+1, -1, 1, 1,
	})
	vbo := g.CreateBuffer()
	g.BindBuffer(gl.ARRAY_BUFFER, vbo)
	g.BufferData(gl.ARRAY_BUFFER, len(vertices), gl.STATIC_DRAW, vertices)
	defer g.DeleteBuffer(vbo)

	program, err := gl.CreateProgram(g, vert, frag, []string{"position", "uvcoord"})
	if err != nil {
		panic(fmt.Sprintf("gles: cannot create shader program: %v", err))
	}
	g.UseProgram(program)
	defer g.DeleteProgram(program)

	position := g.GetAttribLocation(program, "position")
	uvcoord := g.GetAttribLocation(program, "uvcoord")
	g.EnableVertexAttribArray(position)
	g.EnableVertexAttribArray(uvcoord)
	g.VertexAttribPointer(position, 2, gl.FLOAT, false, 4*4, 0)
	g.VertexAttribPointer(uvcoord, 2, gl.FLOAT, false, 4*4, 2*4)

	tex := g.CreateTexture()
	g.BindTexture(gl.TEXTURE_2D, tex)
	defer g.DeleteTexture(tex)

	last := time.Now()
	tPerFrame := time.Second / 240 // 120 fps
	tk := time.NewTicker(tPerFrame)
	for {
		select {
		case siz := <-w.resize:
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
			if a, ok := app.(ResizeHandler); ok {
				a.OnResize(w.win.config.size.X, w.win.config.size.Y)
				continue
			}
		case <-tk.C:
			appdraw, ok := app.(DrawHandler)
			if !ok {
				continue
			}

			img, reDraw := appdraw.Draw()
			if !reDraw {
				continue
			}

			c := time.Now()
			t := c.Sub(last)
			last = c
			if t < tPerFrame {
				continue
			}
			if w.win.config.fps {
				w.fontDrawer.Dot = math.P(5, 15)
				w.fontDrawer.Dst = img
				w.fontDrawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}

			w.flush(img)
			if err := w.win.ctx.Present(); err != nil {
				panic(fmt.Errorf("egl: swap buffer failed: %v", err))
			}
		}
	}
}

// flush uploads the rendered image to a texture and draws it as a full-screen
// quad. The given image is assumed to be a non-nil pointer.
func (w *window) flush(img *image.RGBA) {
	g := w.win.ctx.gl
	g.Viewport(0, 0, w.win.config.size.X, w.win.config.size.Y)
	g.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.Bounds().Dx(), img.Bounds().Dy(), gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)
	g.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	g.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	g.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	g.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	g.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	g.Finish()
}

func slice2bytes(s any) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}

const (
	vert = `#version 100
precision highp float;

attribute vec2 position;
attribute vec2 uvcoord;

varying vec2 outUV;

void main() {
	outUV = uvcoord;
	gl_Position = vec4(position, 0.0, 1.0);
}`
	frag = `#version 100
precision highp float;

varying vec2 outUV;
uniform sampler2D tex;

void main() {
	gl_FragColor = texture2D(tex, outUV);
}`
)

func windowProc(hwnd syscall.Handle, msg uint32, wParam, lParam uintptr) uintptr {
	win, exists := winMap.Load(hwnd)
	if !exists {
		return windows.DefWindowProc(hwnd, msg, wParam, lParam)
	}

	ww := win.(winapp)
	w := ww.win

	switch msg {
	case windows.WM_UNICHAR:
		if wParam == windows.UNICODE_NOCHAR {
			// Tell the system that we accept WM_UNICHAR messages.
			return windows.TRUE
		}
		fallthrough
	case windows.WM_CHAR:
		println("WM_CHAR")
		return windows.TRUE
	case windows.WM_DPICHANGED:
		// Let Windows know we're prepared for runtime DPI changes.
		println("WM_DPICHANGED")
		return windows.TRUE
	case windows.WM_ERASEBKGND:
		// Avoid flickering between GPU content and background color.
		println("WM_ERASEBKGND")
		return windows.TRUE
	case windows.WM_KEYDOWN, windows.WM_KEYUP, windows.WM_SYSKEYDOWN, windows.WM_SYSKEYUP:
		println("WM_KEYDOWN, windows.WM_KEYUP, windows.WM_SYSKEYDOWN, windows.WM_SYSKEYUP")
	case windows.WM_LBUTTONDOWN:
		println("WM_LBUTTONDOWN")
	case windows.WM_LBUTTONUP:
		println("WM_LBUTTONUP")
	case windows.WM_RBUTTONDOWN:
		println("WM_RBUTTONDOWN")
	case windows.WM_RBUTTONUP:
		println("WM_RBUTTONUP")
	case windows.WM_MBUTTONDOWN:
		println("WM_MBUTTONDOWN")
	case windows.WM_MBUTTONUP:
		println("WM_MBUTTONUP")
	case windows.WM_CANCELMODE:
		println("WM_CANCELMODE")
	case windows.WM_SETFOCUS:
		println("WM_SETFOCUS")
	case windows.WM_KILLFOCUS:
		println("WM_KILLFOCUS")
	case windows.WM_MOUSEMOVE:
		println("WM_MOUSEMOVE")
	case windows.WM_MOUSEWHEEL:
		println("WM_MOUSEWHEEL")
	case windows.WM_MOUSEHWHEEL:
		println("WM_MOUSEHWHEEL")
	case windows.WM_DESTROY:
		if w.win.hdc != 0 {
			windows.ReleaseDC(w.win.hdc)
			w.win.hdc = 0
		}
		w.win.hwnd = 0 // The system destroys the HWND for us.
		windows.PostQuitMessage(0)
	case windows.WM_PAINT:
		// Rendering is driven by the background draw goroutine started in run().
		// Fall through to DefWindowProc, which validates the update region so
		// Windows stops re-posting WM_PAINT.

	case windows.WM_SIZE:
		println("WM_SIZE")
	case windows.WM_GETMINMAXINFO:
		mm := (*windows.MinMaxInfo)(unsafe.Pointer(uintptr(lParam)))
		if w.win.config.minSize.X > 0 || w.win.config.maxSize.Y > 0 {
			mm.PtMinTrackSize = windows.Point{
				X: int32(w.win.config.minSize.X),
				Y: int32(w.win.config.minSize.Y),
			}
		}
		if w.win.config.maxSize.X > 0 || w.win.config.maxSize.Y > 0 {
			mm.PtMaxTrackSize = windows.Point{
				X: int32(w.win.config.maxSize.X),
				Y: int32(w.win.config.maxSize.Y),
			}
		}
	case windows.WM_SETCURSOR:
		if (lParam & 0xffff) == windows.HTCLIENT {
			windows.SetCursor(resources.cursor)
			return windows.TRUE
		}
	case windows.WM_USER:
		println("WM_USER")
	}

	return windows.DefWindowProc(hwnd, msg, wParam, lParam)
}
