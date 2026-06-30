// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package render

import (
	"testing"

	"poly.red/math"
)

// TestInterpWorldPos pins the fix for the drawClipped worldpos bug. The three
// vertices lie on the plane x+y+z=10, so any barycentric interpolation of them
// must also sum to 10. The old bug collapsed worldpos to (m1.X, m2.Y, m3.Z) =
// (10,10,10), which sums to 30 -- off the triangle entirely.
func TestInterpWorldPos(t *testing.T) {
	m1 := math.Vec4[float32]{X: 10}
	m2 := math.Vec4[float32]{Y: 10}
	m3 := math.Vec4[float32]{Z: 10}

	c := interpWorldPos([3]float32{1.0 / 3, 1.0 / 3, 1.0 / 3}, m1, m2, m3)
	if sum := c.X + c.Y + c.Z; sum < 9.99 || sum > 10.01 {
		t.Fatalf("centroid worldpos = %v (sum %.3f); want a point on the triangle (sum 10). The bug gave (10,10,10), sum 30", c, sum)
	}

	// At a vertex the interpolation must return that vertex's position (W is the
	// position's homogeneous 1, not compared).
	if v := interpWorldPos([3]float32{1, 0, 0}, m1, m2, m3); v.X != m1.X || v.Y != m1.Y || v.Z != m1.Z {
		t.Fatalf("bc=[1,0,0] -> %v, want %v position", v, m1)
	}
}
