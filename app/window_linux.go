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
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/rand"
	"reflect"
	"runtime"
	"time"
	"unsafe"

	"poly.red/app/internal/gles"
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
	// Make sure all X11 and EGL APIs are called from the same thread.
	runtime.LockOSThread()

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

const vertexShader = `#version 100
uniform vec2 offset;
attribute vec4 position;
void main() {
	// offset comes in with x/y values between 0 and 1.
	// position bounds are -1 to 1.
	vec4 offset4 = vec4(2.0*offset.x-1.0, 1.0-2.0*offset.y, 0, 0);
	gl_Position = position + offset4;
}`

const fragmentShader = `#version 100
precision mediump float;
uniform vec4 color;
void main() {
	gl_FragColor = color;
}`

var triangleData = f32Bytes(binary.LittleEndian,
	0.0, 0.4, 0.0, // top left
	0.0, 0.0, 0.0, // bottom left
	0.4, 0.0, 0.0, // bottom right
)

func f32Bytes(byteOrder binary.ByteOrder, values ...float32) []byte {
	le := false
	switch byteOrder {
	case binary.BigEndian:
	case binary.LittleEndian:
		le = true
	default:
		panic(fmt.Sprintf("invalid byte order %v", byteOrder))
	}

	b := make([]byte, 4*len(values))
	for i, v := range values {
		u := math.Float32bits(v)
		if le {
			b[4*i+0] = byte(u >> 0)
			b[4*i+1] = byte(u >> 8)
			b[4*i+2] = byte(u >> 16)
			b[4*i+3] = byte(u >> 24)
		} else {
			b[4*i+0] = byte(u >> 24)
			b[4*i+1] = byte(u >> 16)
			b[4*i+2] = byte(u >> 8)
			b[4*i+3] = byte(u >> 0)
		}
	}
	return b
}

const (
	coordsPerVertex = 3
	vertexCount     = 3
)

var (
	program  gles.Program
	position gles.Attrib
	offset   gles.Uniform
	col      gles.Uniform
	buf      gles.Buffer

	green  float32
	touchX float32
	touchY float32
)

func slice2bytes(s interface{}) []byte {
	v := reflect.ValueOf(s)
	first := v.Index(0)
	sz := int(first.Type().Size())
	res := unsafe.Slice((*byte)(unsafe.Pointer(v.Pointer())), sz*v.Cap())
	return res[:sz*v.Len()]
}
func (w *window) draw(app Window) {
	// Make sure the drawing calls are always on the same thread.
	runtime.LockOSThread()
	w.win.ctx.Lock()
	defer w.win.ctx.Unlock()

	var err error
	program, err = gles.CreateProgram(w.win.ctx.gl, vertexShader, fragmentShader, []string{
		"position", "color", "offset",
	})
	if err != nil {
		log.Printf("error creating GL program: %v", err)
		return
	}

	buf = w.win.ctx.gl.CreateBuffer()
	w.win.ctx.gl.BindBuffer(gles.ARRAY_BUFFER, buf)
	w.win.ctx.gl.BufferData(gles.ARRAY_BUFFER, len(triangleData), gles.STATIC_DRAW, slice2bytes(triangleData))

	position = w.win.ctx.gl.GetAttribLocation(program, "position")
	col = w.win.ctx.gl.GetUniformLocation(program, "color")
	offset = w.win.ctx.gl.GetUniformLocation(program, "offset")

	tk := time.NewTicker(time.Second / 240) // 120 fps
	for {
		select {
		case siz := <-w.resize:
			if siz.w != w.win.config.size.X && siz.h != w.win.config.size.Y {
				w.win.config.size.X = siz.w
				w.win.config.size.Y = siz.h
				a, ok := app.(ResizeHandler)
				if !ok {
					continue
				}
				a.OnResize(siz.w, siz.h)
			}
		case <-tk.C:
			appdraw, ok := app.(DrawHandler)
			if !ok {
				continue
			}
			img, redraw := appdraw.Draw()
			if redraw {
				continue
			}
			w.flush(frame{img: img})
		case <-w.win.closed:
			w.win.terminate <- event{}
			return
		}

		w.win.ctx.gl.ClearColor(1, 0, 0, 1)
		w.win.ctx.gl.Clear(gles.COLOR_BUFFER_BIT)

		w.win.ctx.gl.UseProgram(program)

		green += 0.01
		if green > 1 {
			green = 0
		}
		w.win.ctx.gl.Uniform4f(col, 0, green, 0, 1)
		w.win.ctx.gl.Uniform2f(offset, rand.Float32(), rand.Float32())
		w.win.ctx.gl.BindBuffer(gles.ARRAY_BUFFER, buf)
		w.win.ctx.gl.EnableVertexAttribArray(position)
		w.win.ctx.gl.VertexAttribPointer(position, coordsPerVertex, gles.FLOAT, false, 0, 0)
		w.win.ctx.gl.DrawArrays(gles.TRIANGLES, 0, vertexCount)
		w.win.ctx.gl.DisableVertexAttribArray(position)

		if err := w.win.ctx.Present(); err != nil {
			w.win.terminate <- event{}
			return
		}
	}
}

// flush flushes the containing pixel buffer of the given image to the
// hardware frame buffer for display prupose. The given image is assumed
// to be non-nil pointer.
func (w *window) flush(f frame) {
	w.win.ctx.gl.Clear(gles.COLOR_BUFFER_BIT)
	w.win.ctx.gl.ClearColor(rand.Float32(), rand.Float32(), rand.Float32(), 1)

	// dx, dy := float32(f.img.Bounds().Dx()), float32(f.img.Bounds().Dy())
}

func (w *window) main(app Window) {
	<-w.ready
	runtime.LockOSThread()

	closed := false
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
		case C.ConfigureNotify: // window configuration change
			cevt := (*C.XConfigureEvent)(unsafe.Pointer(&ev))
			siz := resizeEvent{int(cevt.width), int(cevt.height)}
			w.resize <- siz
		case C.Expose: // update
			// redraw only on the last expose event
			if (*C.XExposeEvent)(unsafe.Pointer(&ev)).count == 0 {
				// TODO: redraw?
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
	w.win.ctx.Release()
	C.XDestroyWindow(w.win.display, w.win.oswin)
	C.XCloseDisplay(w.win.display)
}
