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
	"time"
	"unsafe"

	"github.com/ebitengine/purego"

	"poly.red/gpu/gl"
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
	ctx     *x11Context
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

	// EGL context must be created after the window is created.
	var err error
	w.win.ctx, err = newX11EGLContext(w.win)
	if err != nil {
		panic(fmt.Sprintf("egl: cannot create EGL context for x11: %v", err))
	}
	err = w.win.ctx.Refresh()
	if err != nil {
		panic(fmt.Sprintf("egl: cannot create EGL surface: %v", err))
	}

	go w.draw(app)
	w.ready <- event{}
}

func slice2bytes(s any) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}

func (w *window) draw(app Window) {
	defer func() { w.win.terminate <- event{} }()

	// Make sure the drawing calls are always on the same thread.
	runtime.LockOSThread()
	w.win.ctx.Lock()
	defer w.win.ctx.Unlock()

	vertices := slice2bytes([]float32{
		-1, +1, 0, 0,
		+1, +1, 1, 0,
		-1, -1, 0, 1,
		+1, -1, 1, 1,
	})
	vbo := w.win.ctx.gl.CreateBuffer()
	w.win.ctx.gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	w.win.ctx.gl.BufferData(gl.ARRAY_BUFFER, len(vertices), gl.STATIC_DRAW, vertices)
	defer w.win.ctx.gl.DeleteBuffer(vbo)

	program, err := gl.CreateProgram(w.win.ctx.gl, vert, frag, []string{"position", "uvcoord"})
	if err != nil {
		panic(fmt.Sprintf("gles: cannot creating shader program: %v", err))
	}

	w.win.ctx.gl.UseProgram(program)
	defer w.win.ctx.gl.DeleteProgram(program)

	position := w.win.ctx.gl.GetAttribLocation(program, "position")
	uvcoord := w.win.ctx.gl.GetAttribLocation(program, "uvcoord")

	w.win.ctx.gl.EnableVertexAttribArray(position)
	w.win.ctx.gl.EnableVertexAttribArray(uvcoord)

	w.win.ctx.gl.VertexAttribPointer(position, 2, gl.FLOAT, false, 4*4, 0)
	w.win.ctx.gl.VertexAttribPointer(uvcoord, 2, gl.FLOAT, false, 4*4, 2*4)

	tex := w.win.ctx.gl.CreateTexture()
	w.win.ctx.gl.BindTexture(gl.TEXTURE_2D, tex)
	defer w.win.ctx.gl.DeleteTexture(tex)

	last := time.Now()
	tPerFrame := time.Second / 240 // 120 fps
	tk := time.NewTicker(tPerFrame)
	terminate := false
	for !terminate {
		select {
		case siz := <-w.resize:
			// FIXME: known issue: resizing somehow can cause the GL calls
			// to freeze the entire application. This may only happen on
			// some of drivers.
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
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

			if w.win.config.fps {
				w.fontDrawer.Dot = math.P(5, 15)
				w.fontDrawer.Dst = img
				fps := fmt.Sprintf("%d", time.Second/e.Sub(s))
				w.fontDrawer.DrawString(fps)
			}
			w.flush(img)
			if err := w.win.ctx.Present(); err != nil {
				panic(fmt.Errorf("egl: swap buffer failed: %v", err))
			}
		case <-w.win.closed:
			terminate = true
		}
	}
}

func (w *window) flush(img *image.RGBA) {
	w.win.ctx.gl.Viewport(0, 0, w.win.config.size.X, w.win.config.size.Y)
	w.win.ctx.gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, img.Bounds().Dx(), img.Bounds().Dy(), gl.RGBA, gl.UNSIGNED_BYTE, img.Pix)
	w.win.ctx.gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	w.win.ctx.gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	w.win.ctx.gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	w.win.ctx.gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	w.win.ctx.gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	w.win.ctx.gl.Finish()
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
	w.win.ctx.Release()
	purego.SyscallN(_XDestroyWindow, w.win.display, uintptr(w.win.oswin))
	purego.SyscallN(_XCloseDisplay, w.win.display)
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
