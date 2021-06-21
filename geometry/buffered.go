// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

type BufferedMesh struct {
	positions []float64
	normals   []float64
	uvs       []float64
	color     []float64
	vbo       []int64
}
