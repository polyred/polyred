// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

type Quaternion struct {
	A float64
	V Vec4
}

func NewQuaternion(a, b, c, d float64) Quaternion {
	return Quaternion{
		A: a,
		V: Vec4{b, c, d, 0},
	}
}

func (q Quaternion) Mul(p Quaternion) Quaternion {
	aa := q.A*p.A - q.V.Dot(p.V)
	vv := p.V.Scale(q.A, q.A, q.A, q.A).Add(q.V.Scale(p.A, p.A, p.A, p.A)).Add(q.V.Cross(p.V))
	return Quaternion{aa, vv}
}

func (q Quaternion) ToRoMat() Mat4 {
	w := q.A
	x := q.V.X
	y := q.V.Y
	z := q.V.Z
	m := Mat4{
		1 - 2*y*y - 2*z*z, 2*x*y - 2*z*w, 2*x*z + 2*y*w, 0,
		2*x*y + 2*z*w, 1 - 2*x*x - 2*z*z, 2*y*z - 2*x*w, 0,
		2*x*z - 2*y*w, 2*y*z + 2*x*w, 1 - 2*x*x - 2*y*y, 0,
		0, 0, 0, 1,
	}
	return m
}
