package math

type Quaternion struct {
	A float64
	V Vector
}

func NewQuaternion(a, b, c, d float64) Quaternion {
	return Quaternion{
		A: a,
		V: Vector{b, c, d, 0},
	}
}

func (q Quaternion) Mul(p Quaternion) Quaternion {
	aa := q.A*p.A - q.V.Dot(p.V)
	vv := p.V.Scale(q.A, q.A, q.A, q.A).Add(q.V.Scale(p.A, p.A, p.A, p.A)).Add(q.V.Cross(p.V))
	return Quaternion{aa, vv}
}

func (q Quaternion) ToRoMat() Matrix {
	w := q.A
	x := q.V.X
	y := q.V.Y
	z := q.V.Z
	m := Matrix{
		1 - 2*y*y - 2*z*z, 2*x*y - 2*z*w, 2*x*z + 2*y*w, 0,
		2*x*y + 2*z*w, 1 - 2*x*x - 2*z*z, 2*y*z - 2*x*w, 0,
		2*x*z - 2*y*w, 2*y*z + 2*x*w, 1 - 2*x*x - 2*y*y, 0,
		0, 0, 0, 1,
	}
	return m
}
