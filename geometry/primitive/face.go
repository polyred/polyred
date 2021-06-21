// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

type Face interface {
	AABB() AABB
	Vertices(func(v *Vertex) bool)
	Triangles(func(t *Triangle) bool)
}
