// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package geometry

import (
	"changkun.de/x/polyred/geometry/primitive"
	"changkun.de/x/polyred/math"
)

type BezierCurve struct {
	controlPoints []primitive.Vertex
}

func NewBezierCurve(cp ...*primitive.Vertex) *BezierCurve {
	bc := &BezierCurve{
		controlPoints: make([]primitive.Vertex, len(cp)),
	}
	for i := range cp {
		bc.controlPoints[i] = *cp[i]
	}
	return bc
}

func (bc *BezierCurve) At(t float64) math.Vec4 {
	n := len(bc.controlPoints)

	tc := make([]math.Vec4, n)
	for i := range bc.controlPoints {
		tc[i] = bc.controlPoints[i].Pos
	}

	// The de Casteljau algorithm.
	for j := 0; j < n; j++ {
		for i := 0; i < n-j-1; i++ {
			b01 := math.LerpVec4(tc[i], tc[i+1], t)
			tc[i].X = b01.X
			tc[i].Y = b01.Y
		}
	}
	return tc[0]
}
