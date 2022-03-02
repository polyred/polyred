// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.
//
// Modified from https://github.com/gioui/gio/blob/cdb288d1f98a50b377cc6e916edabb297977670d/app/egl_x11.go

package app

import (
	"poly.red/app/internal/egl"
)

type x11Context struct {
	win *osWindow
	ctx *egl.Context
}

func newX11EGLContext(w *osWindow) (*x11Context, error) {
	ctx, err := egl.NewContext(egl.NativeDisplayType(w.display))
	if err != nil {
		return nil, err
	}
	return &x11Context{win: w, ctx: ctx}, nil
}

func (c *x11Context) Release() {
	if c.ctx != nil {
		c.ctx.Release()
		c.ctx = nil
	}
}

func (c *x11Context) Refresh() error {
	c.ctx.ReleaseSurface()
	eglSurf := egl.NativeWindowType(uintptr(c.win.oswin))
	if err := c.ctx.CreateSurface(eglSurf, c.win.config.size.X, c.win.config.size.Y); err != nil {
		return err
	}
	if err := c.ctx.MakeCurrent(); err != nil {
		return err
	}
	c.ctx.EnableVSync(true)
	c.ctx.ReleaseCurrent()
	return nil
}

func (c *x11Context) Lock() error {
	return c.ctx.MakeCurrent()
}

func (c *x11Context) Unlock() {
	c.ctx.ReleaseCurrent()
}

func (c *x11Context) Present() error {
	return c.ctx.Present()
}
