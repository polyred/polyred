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
	"time"
	"unsafe"

	"poly.red/app/internal/gl"
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
func (w *window) run(app Window, cfg config, opts ...Opt) {
	w.win = &osWindow{
		config:    &cfg,
		closed:    make(chan struct{}, 2),
		terminate: make(chan struct{}, 2),
	}
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

	C.XMapWindow(w.win.display, w.win.oswin)
	C.XClearWindow(w.win.display, w.win.oswin)

	// EGL context must be created after the window is created.
	var err error
	w.win.ctx, err = newX11EGLContext(w.win)
	if err != nil {
		panic("egl: cannot create EGL context for x11")
	}

	go w.event(app)
	go w.draw(app)
	w.ready <- event{}
}

func (w *window) event(app Window) {
	tk := time.NewTicker(time.Second / 960)
	for range tk.C {
		select {
		case key := <-w.keyboard:
			a, ok := app.(KeyboardHalder)
			if !ok {
				continue
			}
			a.OnKey(key)
		case mo := <-w.mouse:
			a, ok := app.(MouseHandler)
			if !ok {
				continue
			}
			a.OnMouse(mo)
		case <-w.win.closed:
			w.win.terminate <- event{}
			return
		}
	}
}

func (w *window) draw(app Window) {
	w.win.ctx.Lock()
	defer w.win.ctx.Unlock()

	// FIXME: not sure why this is not working.
	gl.DrawBuffer(gl.FRONT)
	gl.PixelZoom(1, -1)
	gl.ClearColor(1, 0.5, 1, 1)

	// Managing 3 drawable frames:
	// frames contain their own done indicator, make sure
	// that each frame is indeed drawed from the GPU level.
	frames := [3]frame{
		{done: make(chan event, 1)},
		{done: make(chan event, 1)},
		{done: make(chan event, 1)},
	}
	for i := 0; i < len(frames); i++ {
		frames[i].done <- event{}
	}
	frameIdx := 0

	last := time.Now()
	tPerFrame := time.Second / 240 // 120 fps
	tk := time.NewTicker(tPerFrame)
	for {
		select {
		case siz := <-w.resize:
			// FIXME: notify configs
			// w.win.ctx.Resize(siz.w, siz.h)
			w.win.config.size.X = siz.w
			w.win.config.size.Y = siz.h
			if a, ok := app.(ResizeHandler); ok {
				a.OnResize(siz.w, siz.h)
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

			f := frames[frameIdx]
			f.img = img
			w.flush(f)
			frameIdx = (frameIdx + 1) % 3
		case <-w.win.closed:
			w.win.terminate <- event{}
			return
		}
	}
}

// flush flushes the containing pixel buffer of the given image to the
// hardware frame buffer for display prupose. The given image is assumed
// to be non-nil pointer.
func (w *window) flush(f frame) {
	dx, dy := int32(f.img.Bounds().Dx()), int32(f.img.Bounds().Dy())
	gl.RasterPos2d(-1, 1)
	gl.Viewport(0, 0, dx, dy)
	gl.DrawPixels(dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, f.img.Pix)
	gl.Finish()
}

func (w *window) main(app Window) {
	<-w.ready

	closed := false
	ev := C.XEvent{}
	for !closed {
		C.XNextEvent(w.win.display, &ev)
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
			w.keyboard <- ke
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
			}
			const scrollScale = 10
			switch bevt.button {
			case C.Button1:
				mev.Button = MouseBtnLeft
			case C.Button2:
				mev.Button = MouseBtnMiddle
			case C.Button3:
				mev.Button = MouseBtnRight
			case C.Button4:
				// scroll up
				mev.Action = MouseScroll
				mev.Yoffset = -scrollScale
			case C.Button5:
				// scroll down
				mev.Action = MouseScroll
				mev.Yoffset = +scrollScale
			case 6:
				// http://xahlee.info/linux/linux_x11_mouse_button_number.html
				// scroll left
				mev.Action = MouseScroll
				mev.Xoffset = -scrollScale * 2
			case 7:
				// scroll right
				mev.Action = MouseScroll
				mev.Xoffset = +scrollScale * 2
			}
			w.mouse <- mev
		case C.MotionNotify:
			mevt := (*C.XMotionEvent)(unsafe.Pointer(&ev))
			mev := MouseEvent{
				Action: MouseMove,
				Mods:   ModifierKey(mevt.state),
				Xpos:   float32(mevt.x),
				Ypos:   float32(mevt.y),
			}
			w.mouse <- mev
		case C.ConfigureNotify:
			cevt := (*C.XConfigureEvent)(unsafe.Pointer(&ev))
			w.resize <- resizeEvent{
				int(cevt.width),
				int(cevt.height),
			}
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
	w.win.closed <- event{}
	<-w.win.terminate
	<-w.win.terminate

	// Close the window gracefully.
	C.XDestroyWindow(w.win.display, w.win.oswin)
	C.XCloseDisplay(w.win.display)
}
