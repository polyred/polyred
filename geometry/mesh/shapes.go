// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
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
func NewPlane(width, height float32) Mesh {
	v1 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec4(0, 1, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{255, 0, 0, 255},
	}
	v2 := primitive.Vertex{
		Pos: math.NewVec4(-0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec4(0, 0, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 255, 0, 255},
	}
	v3 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, 0.5*height, 1),
		UV:  math.NewVec4(1, 0, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 255, 255},
	}
	v4 := primitive.Vertex{
		Pos: math.NewVec4(0.5*width, 0, -0.5*height, 1),
		UV:  math.NewVec4(1, 1, 0, 1),
		Nor: math.NewVec4(0, 1, 0, 0),
		Col: color.RGBA{0, 0, 0, 255},
	}
	return NewTriangleSoup([]*primitive.Triangle{
		{V1: v1, V2: v2, V3: v3},
		{V1: v1, V2: v3, V3: v4},
	})
}

// NewRandomTriangleSoup returns a mesh with given number of
// random triangles.
func NewRandomTriangleSoup(numTri int) Mesh {
	vertIdx := make([]uint64, numTri*3)
	vertPos := make([]float32, numTri*3*3)
	vertNor := make([]float32, numTri*3*3)
	vertCol := make([]float32, numTri*3*4)
	vertUV := make([]float32, numTri*3*2)

	for vid := uint64(0); vid < uint64(numTri)*3; vid++ {
		vertIdx[vid] = vid

		vertPos[3*vid] = rand.Float32()
		vertPos[3*vid+1] = rand.Float32()
		vertPos[3*vid+2] = rand.Float32()

		n := math.NewVec3(rand.Float32()*2-1, rand.Float32()*2-1, rand.Float32()*2-1).Unit()
		vertNor[3*vid] = n.X
		vertNor[3*vid+1] = n.Y
		vertNor[3*vid+2] = n.Z

		vertCol[4*vid] = rand.Float32()
		vertCol[4*vid+1] = rand.Float32()
		vertCol[4*vid+2] = rand.Float32()
		vertCol[4*vid+3] = 1

		vertUV[2*vid] = rand.Float32()
		vertUV[2*vid+1] = rand.Float32()
	}

	bm := NewBufferedMesh()
	bm.SetVertexIndex(vertIdx)
	bm.SetAttribute(AttributePos, NewBufferAttribute(3, vertPos))
	bm.SetAttribute(AttributeNor, NewBufferAttribute(3, vertNor))
	bm.SetAttribute(AttributeCol, NewBufferAttribute(4, vertCol))
	bm.SetAttribute(AttributeUV, NewBufferAttribute(2, vertUV))
	return bm
}
