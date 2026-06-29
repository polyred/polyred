// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"github.com/ebitengine/purego"

	"poly.red/gpu"
	"poly.red/math"
)

// X11 constants (values from X11/X.h).
const (
	// Event masks.
	xExposureMask        = 1 << 15
	xKeyPressMask        = 1 << 0
	xKeyReleaseMask      = 1 << 1
	xButtonPressMask     = 1 << 2
	xButtonReleaseMask   = 1 << 3
	xPointerMotionMask   = 1 << 6
	xStructureNotifyMask = 1 << 17
	xFocusChangeMask     = 1 << 21

	// Window attribute valuemask bits.
	xCWBackPixmap       = 1 << 0
	xCWOverrideRedirect = 1 << 9
	xCWEventMask        = 1 << 11

	// Misc.
	xNone           = 0
	xCopyFromParent = 0
	xInputOutput    = 1
	xFalse          = 0
	xTrue           = 1

	// Event type ints.
	xKeyPress        = 2
	xKeyRelease      = 3
	xButtonPress     = 4
	xButtonRelease   = 5
	xMotionNotify    = 6
	xConfigureNotify = 22
	xClientMessage   = 33

	// Button consts.
	xButton1 = 1
	xButton2 = 2
	xButton3 = 3
	xButton4 = 4
	xButton5 = 5
)

// Resolved libX11 entry points (purego function pointers).
var (
	_XInitThreads       uintptr
	_XOpenDisplay       uintptr
	_XCloseDisplay      uintptr
	_XDefaultRootWindow uintptr
	_XCreateWindow      uintptr
	_XInternAtom        uintptr
	_XSetWMProtocols    uintptr
	_XStoreName         uintptr
	_XSetTextProperty   uintptr
	_XMapWindow         uintptr
	_XClearWindow       uintptr
	_XNextEvent         uintptr
	_XFilterEvent       uintptr
	_XDestroyWindow     uintptr
)

var x11LoadOnce sync.Once

func loadX11() error {
	var err error
	x11LoadOnce.Do(func() {
		err = loadX11Symbols()
	})
	return err
}

func loadX11Symbols() error {
	h, err := purego.Dlopen("libX11.so.6", purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		h, err = purego.Dlopen("libX11.so", purego.RTLD_NOW|purego.RTLD_GLOBAL)
		if err != nil {
			return fmt.Errorf("x11: failed to load libX11: %w", err)
		}
	}
	var loadErr error
	sym := func(name string) uintptr {
		p, e := purego.Dlsym(h, name)
		if e != nil && loadErr == nil {
			loadErr = fmt.Errorf("x11: dlsym %s: %w", name, e)
		}
		return p
	}
	_XInitThreads = sym("XInitThreads")
	_XOpenDisplay = sym("XOpenDisplay")
	_XCloseDisplay = sym("XCloseDisplay")
	_XDefaultRootWindow = sym("XDefaultRootWindow")
	_XCreateWindow = sym("XCreateWindow")
	_XInternAtom = sym("XInternAtom")
	_XSetWMProtocols = sym("XSetWMProtocols")
	_XStoreName = sym("XStoreName")
	_XSetTextProperty = sym("XSetTextProperty")
	_XMapWindow = sym("XMapWindow")
	_XClearWindow = sym("XClearWindow")
	_XNextEvent = sym("XNextEvent")
	_XFilterEvent = sym("XFilterEvent")
	_XDestroyWindow = sym("XDestroyWindow")
	return loadErr
}

// cstring returns a NUL-terminated copy of s suitable for passing as a C char*.
func cstring(s string) []byte {
	b := make([]byte, len(s)+1)
	copy(b, s)
	return b
}

// x11SetWindowAttributes mirrors XSetWindowAttributes (X11/Xlib.h) on LP64.
type x11SetWindowAttributes struct {
	backgroundPixmap   uint64 // 0
	backgroundPixel    uint64
	borderPixmap       uint64
	borderPixel        uint64
	bitGravity         int32
	winGravity         int32
	backingStore       int32
	_                  int32 // pad before backingPlanes
	backingPlanes      uint64
	backingPixel       uint64
	saveUnder          int32
	_                  int32 // pad before eventMask
	eventMask          int64 // offset 72
	doNotPropagateMask int64
	overrideRedirect   int32 // offset 88
	_                  int32 // pad before colormap
	colormap           uint64
	cursor             uint64
}

// x11TextProperty mirrors XTextProperty (X11/Xutil.h) on LP64.
type x11TextProperty struct {
	value    uintptr
	encoding uint64
	format   int32
	_        int32
	nitems   uint64
}

// x11KeyEvent mirrors XKeyEvent on LP64.
type x11KeyEvent struct {
	typ        int32
	_          int32
	serial     uint64
	sendEvent  int32
	_          int32
	display    uintptr
	window     uint64
	root       uint64
	subwindow  uint64
	time       uint64
	x, y       int32
	xRoot      int32
	yRoot      int32
	state      uint32
	keycode    uint32 // offset 84
	sameScreen int32
	_          int32
}

// x11ButtonEvent mirrors XButtonEvent on LP64 (identical prefix to key event,
// the field at offset 84 is the button number).
type x11ButtonEvent struct {
	typ        int32
	_          int32
	serial     uint64
	sendEvent  int32
	_          int32
	display    uintptr
	window     uint64
	root       uint64
	subwindow  uint64
	time       uint64
	x, y       int32
	xRoot      int32
	yRoot      int32
	state      uint32
	button     uint32 // offset 84
	sameScreen int32
	_          int32
}

// x11MotionEvent mirrors XMotionEvent on LP64 (we only read x, y, state).
type x11MotionEvent struct {
	typ       int32
	_         int32
	serial    uint64
	sendEvent int32
	_         int32
	display   uintptr
	window    uint64
	root      uint64
	subwindow uint64
	time      uint64
	x, y      int32
	xRoot     int32
	yRoot     int32
	state     uint32
	// is_hint, same_screen unused.
}

// x11ConfigureEvent mirrors XConfigureEvent on LP64 (we only read width/height).
type x11ConfigureEvent struct {
	typ       int32
	_         int32
	serial    uint64
	sendEvent int32
	_         int32
	display   uintptr
	event     uint64
	window    uint64
	x, y      int32
	width     int32
	height    int32
	// (rest unused)
}

// x11ClientMessageEvent mirrors XClientMessageEvent on LP64 (we only read
// data.l[0]).
type x11ClientMessageEvent struct {
	typ         int32
	_           int32
	serial      uint64
	sendEvent   int32
	_           int32
	display     uintptr
	window      uint64
	messageType uint64
	format      int32
	_           int32
	data0       int64 // data.l[0], offset 56
}

type osWindow struct {
	config  *config
	dev     *gpu.Device
	surf    *gpu.Surface
	display uintptr
	oswin   uint64
	atoms   struct {
		utf8string  uint64 // "UTF8_STRING".
		plaintext   uint64 // "text/plain;charset=utf-8".
		wmName      uint64 // "_NET_WM_NAME"
		evDelWindow uint64 // "WM_DELETE_WINDOW"
	}
	closed    chan struct{}
	terminate chan struct{}
}

func (w *window) atom(name string, onlyIfExists bool) uint64 {
	cname := cstring(name)
	var flag uintptr = xFalse
	if onlyIfExists {
		flag = xTrue
	}
	r, _, _ := purego.SyscallN(_XInternAtom, w.win.display, uintptr(unsafe.Pointer(&cname[0])), flag)
	runtime.KeepAlive(cname)
	return uint64(r)
}

var x11Threads sync.Once

func (w *window) run(app Window, cfg config, opts ...Option) {
	// Make sure all X11 and EGL APIs are called from the same thread.
	runtime.LockOSThread()

	w.win = &osWindow{
		config:    &cfg,
		closed:    make(chan struct{}, 1),
		terminate: make(chan struct{}, 1),
	}
	for _, o := range opts {
		o(w.win.config)
	}

	if err := loadX11(); err != nil {
		panic(err.Error())
	}

	x11Threads.Do(func() {
		r, _, _ := purego.SyscallN(_XInitThreads)
		if r == 0 {
			panic("x11: threads init failed")
		}
	})

	d, _, _ := purego.SyscallN(_XOpenDisplay, 0)
	w.win.display = uintptr(d)
	if w.win.display == 0 {
		panic("x11: cannot connect to the X server")
	}

	swa := x11SetWindowAttributes{
		eventMask: xExposureMask | xFocusChangeMask | // update
			xKeyPressMask | xKeyReleaseMask | // keyboard
			xButtonPressMask | xButtonReleaseMask | // mouse clicks
			xPointerMotionMask | // mouse movement
			xStructureNotifyMask, // resize
		backgroundPixmap: xNone,
		overrideRedirect: xFalse,
	}

	root, _, _ := purego.SyscallN(_XDefaultRootWindow, w.win.display)
	oswin, _, _ := purego.SyscallN(_XCreateWindow,
		w.win.display,
		root,
		0, 0,
		uintptr(w.win.config.size.X), uintptr(w.win.config.size.Y),
		0, xCopyFromParent, xInputOutput, 0, // border_width, depth, class, visual
		xCWEventMask|xCWBackPixmap|xCWOverrideRedirect,
		uintptr(unsafe.Pointer(&swa)))
	w.win.oswin = uint64(oswin)
	runtime.KeepAlive(&swa)

	w.win.atoms.utf8string = w.atom("UTF8_STRING", false)
	w.win.atoms.plaintext = w.atom("text/plain;charset=utf-8", false)
	w.win.atoms.wmName = w.atom("_NET_WM_NAME", false)
	w.win.atoms.evDelWindow = w.atom("WM_DELETE_WINDOW", false)

	// extensions
	purego.SyscallN(_XSetWMProtocols, w.win.display, uintptr(w.win.oswin),
		uintptr(unsafe.Pointer(&w.win.atoms.evDelWindow)), 1)
	runtime.KeepAlive(&w.win.atoms.evDelWindow)

	ctitle := cstring(w.win.config.title)
	purego.SyscallN(_XStoreName, w.win.display, uintptr(w.win.oswin),
		uintptr(unsafe.Pointer(&ctitle[0])))
	tp := x11TextProperty{
		value:    uintptr(unsafe.Pointer(&ctitle[0])),
		encoding: w.win.atoms.utf8string,
		format:   8,
		nitems:   uint64(len(w.win.config.title)),
	}
	purego.SyscallN(_XSetTextProperty, w.win.display, uintptr(w.win.oswin),
		uintptr(unsafe.Pointer(&tp)), uintptr(w.win.atoms.wmName))
	runtime.KeepAlive(ctitle)
	runtime.KeepAlive(&tp)

	// Let the window to appear.
	purego.SyscallN(_XMapWindow, w.win.display, uintptr(w.win.oswin))
	purego.SyscallN(_XClearWindow, w.win.display, uintptr(w.win.oswin))

	// The GPU device (GL backend) drives windowed present through the Device API:
	// the app uploads each CPU frame to an on-screen Surface bound to this X11
	// window, and the backend blits + swaps it. The window must exist first
	// (eglCreateWindowSurface needs the native window).
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		panic(fmt.Sprintf("gpu: cannot open GL device: %v", err))
	}
	w.win.dev = dev
	surf, err := dev.CreateWindowSurface(gpu.WindowSurfaceDescriptor{
		Display: w.win.display,
		Window:  uintptr(w.win.oswin),
		Width:   w.win.config.size.X,
		Height:  w.win.config.size.Y,
		Format:  gpu.RGBA8Unorm,
	})
	if err != nil {
		panic(fmt.Sprintf("gpu: cannot create window surface: %v", err))
	}
	w.win.surf = surf

	go w.draw(app)
	w.ready <- event{}
}

func (w *window) draw(app Window) {
	defer func() { w.win.terminate <- event{} }()

	last := time.Now()
	tPerFrame := time.Second / 240 // 120 fps
	tk := time.NewTicker(tPerFrame)
	defer tk.Stop()
	terminate := false
	for !terminate {
		select {
		case siz := <-w.resize:
			// FIXME: known issue: resizing somehow can cause the GL calls
			// to freeze the entire application. This may only happen on
			// some of drivers.
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
			if err := w.win.surf.Resize(siz.w, siz.h); err != nil {
				panic(fmt.Sprintf("gpu: surface resize failed: %v", err))
			}
			if a, ok := app.(ResizeHandler); ok {
				a.OnResize(siz.w, siz.h)
			}
		case <-tk.C:
			appdraw, ok := app.(DrawHandler)
			if !ok {
				continue
			}

			s := time.Now()
			img, redraw := appdraw.Draw()
			if !redraw {
				continue
			}

			e := time.Now()
			t := e.Sub(last)
			last = e
			if t < tPerFrame {
				continue
			}

			// Keep the surface sized to the frame so PresentImage's upload
			// matches (the app renders at the window size; this also covers any
			// transient mismatch before a resize event lands).
			if b := img.Bounds(); b.Dx() != w.win.config.size.X || b.Dy() != w.win.config.size.Y {
				w.win.config.size.X, w.win.config.size.Y = b.Dx(), b.Dy()
				if err := w.win.surf.Resize(b.Dx(), b.Dy()); err != nil {
					panic(fmt.Sprintf("gpu: surface resize failed: %v", err))
				}
			}

			if w.win.config.fps {
				w.fontDrawer.Dot = math.P(5, 15)
				w.fontDrawer.Dst = img
				fps := fmt.Sprintf("%d", time.Second/e.Sub(s))
				w.fontDrawer.DrawString(fps)
			}
			if err := w.win.surf.PresentImage(img); err != nil {
				panic(fmt.Sprintf("gpu: present failed: %v", err))
			}
		case <-w.win.closed:
			terminate = true
		}
	}
}

func (w *window) main(app Window) {
	<-w.ready
	runtime.LockOSThread()

	closed := false
	lastButton := MouseBtnNone
	var ev [24]uint64 // XEvent union: long pad[24] (192 bytes).
	for !closed {
		purego.SyscallN(_XNextEvent, w.win.display, uintptr(unsafe.Pointer(&ev[0])))
		if r, _, _ := purego.SyscallN(_XFilterEvent, uintptr(unsafe.Pointer(&ev[0])), xNone); r == xTrue {
			continue
		}

		switch etype := *(*int32)(unsafe.Pointer(&ev[0])); etype {
		case xKeyPress, xKeyRelease:
			ke := KeyEvent{}
			if etype == xKeyPress {
				ke.Pressed = true
			}
			kevt := (*x11KeyEvent)(unsafe.Pointer(&ev[0]))

			ke.Keycode = Key{
				code: uint32(kevt.keycode),
				char: "",
			}
			ke.Mods = x11ModsToLogical(uint32(kevt.state))
			// FIXME: convert keycode to char
			a, ok := app.(KeyboardHanlder)
			if !ok {
				continue
			}
			a.OnKey(ke)
		case xButtonPress, xButtonRelease:
			bevt := (*x11ButtonEvent)(unsafe.Pointer(&ev[0]))
			mev := MouseEvent{
				Action: MouseDown,
				Mods:   x11ModsToLogical(uint32(bevt.state)),
				Xpos:   float32(bevt.x),
				Ypos:   float32(bevt.y),
			}
			if etype == xButtonRelease {
				mev.Action = MouseUp
				lastButton = MouseBtnNone
			}

			switch bevt.button {
			case xButton1:
				mev.Button = MouseBtnLeft
				if etype == xButtonPress {
					lastButton = MouseBtnLeft
				}
			case xButton2:
				if etype == xButtonPress {
					lastButton = MouseBtnMiddle
				}
				mev.Button = MouseBtnMiddle
			case xButton3:
				if etype == xButtonPress {
					lastButton = MouseBtnRight
				}
				mev.Button = MouseBtnRight
			case xButton4:
				// scroll up
				mev.Action = MouseScroll
				mev.Yoffset = -1
			case xButton5:
				// scroll down
				mev.Action = MouseScroll
				mev.Yoffset = +1
			case 6:
				// http://xahlee.info/linux/linux_x11_mouse_button_number.html
				// scroll left
				mev.Action = MouseScroll
				mev.Xoffset = -1
			case 7:
				// scroll right
				mev.Action = MouseScroll
				mev.Xoffset = +1
			}
			a, ok := app.(MouseHandler)
			if !ok {
				continue
			}
			a.OnMouse(mev)
		case xMotionNotify:
			mevt := (*x11MotionEvent)(unsafe.Pointer(&ev[0]))
			mev := MouseEvent{
				Button: lastButton,
				Action: MouseMove,
				Mods:   x11ModsToLogical(uint32(mevt.state)),
				Xpos:   float32(mevt.x),
				Ypos:   float32(mevt.y),
			}
			a, ok := app.(MouseHandler)
			if !ok {
				continue
			}
			a.OnMouse(mev)
		case xConfigureNotify: // window configuration change
			cevt := (*x11ConfigureEvent)(unsafe.Pointer(&ev[0]))
			siz := resizeEvent{w: int(cevt.width), h: int(cevt.height)}
			w.resize <- siz
		case xClientMessage: // extensions
			cevt := (*x11ClientMessageEvent)(unsafe.Pointer(&ev[0]))
			switch uint64(cevt.data0) {
			case w.win.atoms.evDelWindow:
				closed = true
			}
		}
	}

	// Notify and close the event and draw loop.
	w.win.closed <- event{}
	<-w.win.terminate

	// Close the window gracefully.
	w.win.surf.Release()
	w.win.dev.Close()
	purego.SyscallN(_XDestroyWindow, w.win.display, uintptr(w.win.oswin))
	purego.SyscallN(_XCloseDisplay, w.win.display)
}
