// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"poly.red/gpu/ctx/egl"
	"poly.red/gpu/gl"
)

// winContext drives presentation through an EGL (ANGLE) context, mirroring the
// X11 path in ctx_egl_linux.go. ANGLE accepts the window's device context as the
// native display and the window handle as the native window.
type winContext struct {
	win *osWindow
	ctx *egl.Context
	gl  *gl.Functions
}

func newWinContext(w *osWindow) (*winContext, error) {
	ctx, err := egl.NewContext(egl.NativeDisplayType(w.hdc))
	if err != nil {
		return nil, err
	}
	f, err := gl.NewFunctions()
	if err != nil {
		return nil, err
	}
	return &winContext{win: w, ctx: ctx, gl: f}, nil
}

func (c *winContext) Release() {
	if c.ctx != nil {
		c.ctx.Release()
		c.ctx = nil
	}
}

func (c *winContext) Refresh() error {
	c.ctx.ReleaseSurface()
	surf := egl.NativeWindowType(uintptr(c.win.hwnd))
	if err := c.ctx.CreateSurface(surf, c.win.config.size.X, c.win.config.size.Y); err != nil {
		return err
	}
	if err := c.ctx.MakeCurrent(); err != nil {
		return err
	}
	c.ctx.EnableVSync(true)
	c.ctx.ReleaseCurrent()
	return nil
}

func (c *winContext) Lock() error {
	return c.ctx.MakeCurrent()
}

func (c *winContext) Unlock() {
	c.ctx.ReleaseCurrent()
}

func (c *winContext) Present() error {
	return c.ctx.Present()
}
