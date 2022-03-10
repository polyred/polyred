// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package math

type Quaternion[T Float] struct {
	A T
	V Vec3[T]
}

func NewQuaternion[T Float](a, b, c, d T) Quaternion[T] {
	return Quaternion[T]{
		A: a,
		V: Vec3[T]{b, c, d},
	}
}

func (q *Quaternion[T]) Mul(p Quaternion[T]) Quaternion[T] {
	aa := q.A*p.A - q.V.Dot(p.V)
	vv := p.V.Scale(q.A, q.A, q.A).Add(q.V.Scale(p.A, p.A, p.A)).Add(q.V.Cross(p.V))
	return Quaternion[T]{aa, vv}
}

func (q *Quaternion[T]) ToRoMat() Mat4[T] {
	w := q.A
	x := q.V.X
	y := q.V.Y
	z := q.V.Z
	m := Mat4[T]{
		1 - 2*y*y - 2*z*z, 2*x*y - 2*z*w, 2*x*z + 2*y*w, 0,
		2*x*y + 2*z*w, 1 - 2*x*x - 2*z*z, 2*y*z - 2*x*w, 0,
		2*x*z - 2*y*w, 2*y*z + 2*x*w, 1 - 2*x*x - 2*y*y, 0,
		0, 0, 0, 1,
	}
	return m
}
