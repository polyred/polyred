// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

/*
#cgo linux pkg-config: x11

#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
*/
import "C"
import (
	"fmt"
	"image"
	"runtime"
	"sync"
	"time"
	"unsafe"

	"poly.red/internal/bytes"
	"poly.red/internal/driver/gles"
	"poly.red/math"
)

type osWindow struct {
	config  *config
	ctx     *x11Context
	display *C.Display
	oswin   C.Window
	atoms   struct {
		utf8string  C.Atom // "UTF8_STRING".
		plaintext   C.Atom // "text/plain;charset=utf-8".
		wmName      C.Atom // "_NET_WM_NAME"
		evDelWindow C.Atom // "WM_DELETE_WINDOW"
	}
	closed    chan struct{}
	terminate chan struct{}
}

func (w *window) atom(name string, onlyIfExists bool) C.Atom {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	flag := C.Bool(C.False)
	if onlyIfExists {
		flag = C.True
	}
	return C.XInternAtom(w.win.display, cname, flag)
}

var x11Threads sync.Once

func (w *window) run(app Window, cfg config, opts ...Opt) {
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

	x11Threads.Do(func() {
		if C.XInitThreads() == 0 {
			panic("x11: threads init failed")
		}
	})

	w.win.display = C.XOpenDisplay(nil)
	if w.win.display == nil {
		panic("x11: cannot connect to the X server")
	}

	swa := C.XSetWindowAttributes{
		event_mask: C.ExposureMask | C.FocusChangeMask | // update
			C.KeyPressMask | C.KeyReleaseMask | // keyboard
			C.ButtonPressMask | C.ButtonReleaseMask | // mouse clicks
			C.PointerMotionMask | // mouse movement
			C.StructureNotifyMask, // resize
		background_pixmap: C.None,
		override_redirect: C.False,
	}
	w.win.oswin = C.XCreateWindow(w.win.display,
		C.XDefaultRootWindow(w.win.display),
		0, 0, C.uint(w.win.config.size.X), C.uint(w.win.config.size.Y),
		0, C.CopyFromParent, C.InputOutput, nil,
		C.CWEventMask|C.CWBackPixmap|C.CWOverrideRedirect, &swa)

	w.win.atoms.utf8string = w.atom("UTF8_STRING", false)
	w.win.atoms.plaintext = w.atom("text/plain;charset=utf-8", false)
	w.win.atoms.wmName = w.atom("_NET_WM_NAME", false)
	w.win.atoms.evDelWindow = w.atom("WM_DELETE_WINDOW", false)

	// extensions
	C.XSetWMProtocols(w.win.display, w.win.oswin, &w.win.atoms.evDelWindow, 1)

	ctitle := C.CString(w.win.config.title)
	defer C.free(unsafe.Pointer(ctitle))
	C.XStoreName(w.win.display, w.win.oswin, ctitle)
	C.XSetTextProperty(w.win.display, w.win.oswin,
		&C.XTextProperty{
			value:    (*C.uchar)(unsafe.Pointer(ctitle)),
			encoding: w.win.atoms.utf8string,
			format:   8,
			nitems:   C.ulong(len(w.win.config.title)),
		}, w.win.atoms.wmName)

	// Let the window to appear.
	C.XMapWindow(w.win.display, w.win.oswin)
	C.XClearWindow(w.win.display, w.win.oswin)

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

func (w *window) draw(app Window) {
	defer func() { w.win.terminate <- event{} }()

	// Make sure the drawing calls are always on the same thread.
	runtime.LockOSThread()
	w.win.ctx.Lock()
	defer w.win.ctx.Unlock()

	vertices := bytes.FromSlice([]float32{
		-1, +1, 0, 0,
		+1, +1, 1, 0,
		-1, -1, 0, 1,
		+1, -1, 1, 1,
	})
	vbo := w.win.ctx.gl.CreateBuffer()
	w.win.ctx.gl.BindBuffer(gles.ARRAY_BUFFER, vbo)
	w.win.ctx.gl.BufferData(gles.ARRAY_BUFFER, len(vertices), gles.STATIC_DRAW, vertices)
	defer w.win.ctx.gl.DeleteBuffer(vbo)

	program, err := gles.CreateProgram(w.win.ctx.gl, vert, frag, []string{"position", "uvcoord"})
	if err != nil {
		panic(fmt.Sprintf("gles: cannot creating shader program: %v", err))
	}

	w.win.ctx.gl.UseProgram(program)
	defer w.win.ctx.gl.DeleteProgram(program)

	position := w.win.ctx.gl.GetAttribLocation(program, "position")
	uvcoord := w.win.ctx.gl.GetAttribLocation(program, "uvcoord")

	w.win.ctx.gl.EnableVertexAttribArray(position)
	w.win.ctx.gl.EnableVertexAttribArray(uvcoord)

	w.win.ctx.gl.VertexAttribPointer(position, 2, gles.FLOAT, false, 4*4, 0)
	w.win.ctx.gl.VertexAttribPointer(uvcoord, 2, gles.FLOAT, false, 4*4, 2*4)

	tex := w.win.ctx.gl.CreateTexture()
	w.win.ctx.gl.BindTexture(gles.TEXTURE_2D, tex)
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
	w.win.ctx.gl.TexImage2D(gles.TEXTURE_2D, 0, gles.RGBA, img.Bounds().Dx(), img.Bounds().Dy(), gles.RGBA, gles.UNSIGNED_BYTE, img.Pix)
	w.win.ctx.gl.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_WRAP_S, gles.CLAMP_TO_EDGE)
	w.win.ctx.gl.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_WRAP_T, gles.CLAMP_TO_EDGE)
	w.win.ctx.gl.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_MIN_FILTER, gles.LINEAR)
	w.win.ctx.gl.TexParameteri(gles.TEXTURE_2D, gles.TEXTURE_MAG_FILTER, gles.LINEAR)
	w.win.ctx.gl.DrawArrays(gles.TRIANGLE_STRIP, 0, 4)
	w.win.ctx.gl.Finish()
}

func (w *window) main(app Window) {
	<-w.ready
	runtime.LockOSThread()

	closed := false
	lastButton := MouseBtnNone
	ev := C.XEvent{}
	for !closed {
		C.XNextEvent(w.win.display, &ev)
		if C.XFilterEvent(&ev, C.None) == C.True {
			continue
		}

		switch _type := (*C.XAnyEvent)(unsafe.Pointer(&ev))._type; _type {
		case C.KeyPress, C.KeyRelease:
			ke := KeyEvent{}
			if _type == C.KeyPress {
				ke.Pressed = true
			}
			kevt := (*C.XKeyEvent)(unsafe.Pointer(&ev))

			ke.Keycode = Key{
				code: uint32(kevt.keycode),
				char: "",
			}
			ke.Mods = ModifierKey(kevt.state)
			// FIXME: convert keycode to char
			a, ok := app.(KeyboardHanlder)
			if !ok {
				continue
			}
			a.OnKey(ke)
		case C.ButtonPress, C.ButtonRelease:
			bevt := (*C.XButtonEvent)(unsafe.Pointer(&ev))
			mev := MouseEvent{
				Action: MouseDown,
				Mods:   ModifierKey(bevt.state),
				Xpos:   float32(bevt.x),
				Ypos:   float32(bevt.y),
			}
			if bevt._type == C.ButtonRelease {
				mev.Action = MouseUp
				lastButton = MouseBtnNone
			}

			switch bevt.button {
			case C.Button1:
				mev.Button = MouseBtnLeft
				if bevt._type == C.ButtonPress {
					lastButton = MouseBtnLeft
				}
			case C.Button2:
				if bevt._type == C.ButtonPress {
					lastButton = MouseBtnMiddle
				}
				mev.Button = MouseBtnMiddle
			case C.Button3:
				if bevt._type == C.ButtonPress {
					lastButton = MouseBtnRight
				}
				mev.Button = MouseBtnRight
			case C.Button4:
				// scroll up
				mev.Action = MouseScroll
				mev.Yoffset = -1
			case C.Button5:
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
		case C.MotionNotify:
			mevt := (*C.XMotionEvent)(unsafe.Pointer(&ev))
			mev := MouseEvent{
				Button: lastButton,
				Action: MouseMove,
				Mods:   ModifierKey(mevt.state),
				Xpos:   float32(mevt.x),
				Ypos:   float32(mevt.y),
			}
			a, ok := app.(MouseHandler)
			if !ok {
				continue
			}
			a.OnMouse(mev)
		case C.ConfigureNotify: // window configuration change
			cevt := (*C.XConfigureEvent)(unsafe.Pointer(&ev))
			siz := resizeEvent{w: int(cevt.width), h: int(cevt.height)}
			w.resize <- siz
		case C.ClientMessage: // extensions
			cevt := (*C.XClientMessageEvent)(unsafe.Pointer(&ev))
			switch *(*C.long)(unsafe.Pointer(&cevt.data)) {
			case C.long(w.win.atoms.evDelWindow):
				closed = true
			}
		}
	}

	// Notify and close the event and draw loop.
	w.win.closed <- event{}
	<-w.win.terminate

	// Close the window gracefully.
	w.win.ctx.Release()
	C.XDestroyWindow(w.win.display, w.win.oswin)
	C.XCloseDisplay(w.win.display)
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
