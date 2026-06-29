// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/syscall/windows"
	"poly.red/math"
)

type osWindow struct {
	hwnd syscall.Handle
	hdc  syscall.Handle
	dev  *gpu.Device
	surf *gpu.Surface

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

	// Open the GPU device (GL backend, ANGLE on Windows) on the window's device
	// context, then bind an on-screen Surface to the HWND. ANGLE's
	// eglCreateWindowSurface takes the HWND directly, so unlike X11 there is no
	// visual to match. This mirrors the X11 path which builds the device + surface
	// after the window is shown.
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL), gpu.WithNativeDisplay(uintptr(w.win.hdc)))
	if err != nil {
		panic(fmt.Errorf("gpu: cannot open GL device: %w", err))
	}
	w.win.dev = dev
	surf, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Display: uintptr(w.win.hdc), Window: uintptr(w.win.hwnd),
		Width: w.win.config.size.X, Height: w.win.config.size.Y, Format: gpu.RGBA8Unorm,
	})
	if err != nil {
		panic(fmt.Errorf("gpu: cannot create window surface: %w", err))
	}
	w.win.surf = surf

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
	// The GL backend owns its own context thread; locking here mirrors the X11
	// draw goroutine and the Win32 model.
	runtime.LockOSThread()

	last := time.Now()
	tPerFrame := time.Second / 240 // 120 fps
	tk := time.NewTicker(tPerFrame)
	for {
		select {
		case siz := <-w.resize:
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
			if err := w.win.surf.Resize(siz.w, siz.h); err != nil {
				panic(fmt.Errorf("gpu: surface resize failed: %w", err))
			}
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

			// Keep the surface sized to the frame so PresentImage's upload matches
			// (the app renders at the window size; this also covers any transient
			// mismatch before a resize event lands).
			if b := img.Bounds(); b.Dx() != w.win.config.size.X || b.Dy() != w.win.config.size.Y {
				w.win.config.size.X, w.win.config.size.Y = b.Dx(), b.Dy()
				if err := w.win.surf.Resize(b.Dx(), b.Dy()); err != nil {
					panic(fmt.Errorf("gpu: surface resize failed: %w", err))
				}
			}

			if w.win.config.fps {
				w.fontDrawer.Dot = math.P(5, 15)
				w.fontDrawer.Dst = img
				w.fontDrawer.DrawString(fmt.Sprintf("%d", time.Second/t))
			}

			if err := w.win.surf.PresentImage(img); err != nil {
				panic(fmt.Errorf("gpu: present failed: %w", err))
			}
		}
	}
}

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
