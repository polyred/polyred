// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package gl

/*
#cgo CFLAGS: -Werror

#cgo linux LDFLAGS: -lGL

#cgo darwin LDFLAGS: -framework OpenGL
#cgo darwin CFLAGS: -DGL_SILENCE_DEPRECATION

#include <stdlib.h>

#ifdef __APPLE__
	#include "TargetConditionals.h"
	#include <OpenGL/gl.h>
#else
#define __USE_GNU
#include <GL/gl.h>
#endif
*/
import "C"

type (
	Attrib uint
	Enum   uint
)

const (
	FRONT               Enum = 0x0404
	UNSIGNED_BYTE       Enum = 0x1401
	RGBA                Enum = 0x1908
	BGRA                Enum = 0x80E1
	GL_COLOR_BUFFER_BIT Enum = C.GL_COLOR_BUFFER_BIT
	GL_DEPTH_BUFFER_BIT Enum = C.GL_DEPTH_BUFFER_BIT
)
