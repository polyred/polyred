// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package mesh

import (
	"math/rand"

	"poly.red/math"
)

func NewRandomAs[T Mesh](numTri int) (m T) {
	switch any(m).(type) {
	case *BufferedMesh:
		ibo := make([]int, numTri*3)
		vertPos := make([]float32, numTri*3*3)
		vertNor := make([]float32, numTri*3*3)
		vertCol := make([]float32, numTri*3*4)
		vertUV := make([]float32, numTri*3*2)

		for vid := 0; vid < numTri*3; vid++ {
			ibo[vid] = vid

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
		bm.SetIndexBuffer(ibo)
		bm.SetAttribute(AttribPosition, NewBufferAttrib(3, vertPos))
		bm.SetAttribute(AttribNormal, NewBufferAttrib(3, vertNor))
		bm.SetAttribute(AttribColor, NewBufferAttrib(4, vertCol))
		bm.SetAttribute(AttriTexcoord, NewBufferAttrib(2, vertUV))
		return any(bm).(T)
	default:
		panic("not supported")
	}
}
