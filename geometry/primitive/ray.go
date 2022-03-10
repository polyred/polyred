// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package primitive

import "poly.red/math"

// Ray is a ray from Ori position towards Dir direction.
type Ray struct {
	Ori, Dir math.Vec3[float32]
}

func (r *Ray) Pos(t float32) math.Vec3[float32] {
	return r.Ori.Add(r.Dir.Scale(t, t, t))
}
