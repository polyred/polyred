// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

// Face represents a polygon face.
type Face interface {
	Normal() math.Vec3
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}
