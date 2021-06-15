// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import (
	"image/color"

	"changkun.de/x/ddd/math"
)

// Vertex is a vertex that contains the necessary information for
// describing a mesh.
type Vertex struct {
	Pos math.Vector
	UV  math.Vector
	Nor math.Vector
	Col color.RGBA
}

// Triangle is a triangle that contains three vertices.
type Triangle struct {
	V1, V2, V3 Vertex
}
