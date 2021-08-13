// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !darwin
// +build !darwin

package gui

import (
	"fmt"
	"image"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"poly.red/render"
)

type driverInfo struct {
	// Nothing for GL
}

func (w *win) initCallbacks() {
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

func (w *win) initWinHints() {
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.DoubleBuffer, glfw.False)
	glfw.WindowHint(glfw.Resizable, glfw.True)
}

func (w *win) initDriver() {
	w.win.MakeContextCurrent()
	err := gl.Init()
	if err != nil {
		panic(fmt.Errorf("failed to initialize gl: %w", err))
	}
	gl.DrawBuffer(gl.FRONT)
	gl.PixelZoom(1, -1)
}

// flush flushes the containing pixel buffer of the given image to the
// hardware frame buffer for display prupose. The given image is assumed
// to be non-nil pointer.
func (w *win) flush(img *image.RGBA) error {
	dx, dy := int32(img.Bounds().Dx()), int32(img.Bounds().Dy())
	gl.RasterPos2d(-1, 1)
	gl.Viewport(0, 0, dx, dy)
	gl.DrawPixels(dx, dy, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&img.Pix[0]))

	// We need a synchornization here. Similar to commandBuffer.WaitUntilCompleted.
	// See a general discussion about CPU, GPU and display synchornization here:
	//
	// Working with Metal: Fundamentals, 21:28
	// https://developer.apple.com/videos/play/wwdc2014/604/
	//
	// The difference of gl.Finish and gl.Flush can be found here:
	// https://www.khronos.org/registry/OpenGL-Refpages/gl2.1/xhtml/glFlush.xml
	// https://www.khronos.org/registry/OpenGL-Refpages/gl2.1/xhtml/glFinish.xml
	//
	// We may not need such an wait, if we are doing perfect timing.
	// See: https://golang.design/research/ultimate-channel/
	// gl.Finish()
	gl.Flush()
	return nil
}
