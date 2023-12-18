// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"syscall"

	"poly.red/gpu/gl"
)

type glContext struct {
}

func newGLContext(hdc syscall.Handle) (glContext, error) {
	gl.MakeCurrent(hdc)
	gl.DrawBuffer(gl.FRONT)
	gl.PixelZoom(1, -1)
	return glContext{}, nil
}
