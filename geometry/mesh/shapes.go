// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"image/color"
	"math/rand"

	"poly.red/geometry/primitive"
	"poly.red/math"
)

// NewPlane returns a triangle soup that represents a plane with the
// given width and height.
func NewPlane(width, height float64) Mesh {
	v1 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec2(0, 1),
		Nor: math.NewVec3(0, 1, 0),
		Col: color.RGBA{255, 0, 0, 255},
	}
	v2 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec2(0, 0),
		Nor: math.NewVec3(0, 1, 0),
		Col: color.RGBA{0, 255, 0, 255},
	}
	v3 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec2(1, 0),
		Nor: math.NewVec3(0, 1, 0),
		Col: color.RGBA{0, 0, 255, 255},
	}
	v4 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec2(1, 1),
		Nor: math.NewVec3(0, 1, 0),
		Col: color.RGBA{0, 0, 0, 255},
	}
	return NewTriangleSoup([]*primitive.Triangle{
		{V: [3]primitive.Vertex{v1, v2, v3}},
		{V: [3]primitive.Vertex{v1, v3, v4}},
	})
}

// NewRandomTriangleSoup returns a mesh with given number of
// random triangles.
func NewRandomTriangleSoup(numTri int) Mesh {
	idx := make([]uint64, numTri*3)
	pos := make([]float64, numTri*3)
	nor := make([]float64, numTri*3)
	uv := make([]float64, numTri*2)
	col := make([]float64, numTri*3)

	for i := uint64(0); i < uint64(numTri); i++ {
		idx[3*i] = i
		idx[3*i+1] = i + 1
		idx[3*i+2] = i + 2

		pos[3*i] = rand.Float64()*2 - 1
		pos[3*i+1] = rand.Float64()*2 - 1
		pos[3*i+2] = rand.Float64()*2 - 1

		nor[3*i] = rand.Float64()*2 - 1
		nor[3*i+1] = rand.Float64()*2 - 1
		nor[3*i+2] = rand.Float64()*2 - 1

		col[3*i] = math.Clamp(rand.Float64()*0xff, 0, 255)
		col[3*i+1] = math.Clamp(rand.Float64()*0xff, 0, 255)
		col[3*i+2] = math.Clamp(rand.Float64()*0xff, 0, 255)

		uv[2*i] = rand.Float64()*2 - 1
		uv[2*i+1] = rand.Float64()*2 - 1
	}

	bm := NewBufferedMesh()
	bm.SetVertexIndex(idx)
	bm.SetAttr(AttrPos, NewBufferAttr(3, pos))
	bm.SetAttr(AttrNor, NewBufferAttr(3, nor))
	bm.SetAttr(AttrCol, NewBufferAttr(3, col))
	bm.SetAttr(AttrUV, NewBufferAttr(2, uv))
	return bm
}
