// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gpumath

import (
	"math"
	"testing"
)

func almost(a, b float32) bool { return math.Abs(float64(a-b)) < 1e-5 }

func TestVec4Ops(t *testing.T) {
	a := Vec4{1, 2, 3, 4}
	b := Vec4{5, 6, 7, 8}
	if got := a.Add(b); got != (Vec4{6, 8, 10, 12}) {
		t.Errorf("Add = %v", got)
	}
	if got := b.Sub(a); got != (Vec4{4, 4, 4, 4}) {
		t.Errorf("Sub = %v", got)
	}
	if got := a.Mul(b); got != (Vec4{5, 12, 21, 32}) {
		t.Errorf("Mul = %v", got)
	}
	if got := a.Scale(2); got != (Vec4{2, 4, 6, 8}) {
		t.Errorf("Scale = %v", got)
	}
	if got := a.Div(2); got != (Vec4{0.5, 1, 1.5, 2}) {
		t.Errorf("Div = %v", got)
	}
	if got := a.Dot(b); got != 70 {
		t.Errorf("Dot = %v, want 70", got)
	}
	// free functions mirror methods
	if Dot(a, b) != a.Dot(b) || Add(a, b) != a.Add(b) {
		t.Errorf("free functions disagree with methods")
	}
}

func TestNormalize(t *testing.T) {
	v := Vec4{3, 0, 4, 0}
	n := v.Normalize()
	if !almost(n.Length(), 1) {
		t.Errorf("normalized length = %v, want 1", n.Length())
	}
	if !almost(n.X, 0.6) || !almost(n.Z, 0.8) {
		t.Errorf("normalize = %v, want (0.6,0,0.8,0)", n)
	}
	if (Vec4{}).Normalize() != (Vec4{}) {
		t.Errorf("normalize of zero must not divide by zero")
	}
}

func TestMat4MulV(t *testing.T) {
	// Column-major identity: M*v == v.
	id := Mat4{Vec4{1, 0, 0, 0}, Vec4{0, 1, 0, 0}, Vec4{0, 0, 1, 0}, Vec4{0, 0, 0, 1}}
	v := Vec4{2, 3, 4, 1}
	if got := id.MulV(v); got != v {
		t.Errorf("identity MulV = %v, want %v", got, v)
	}
	// A scaling matrix (columns): diag(2,3,4,1).
	s := Mat4{Vec4{2, 0, 0, 0}, Vec4{0, 3, 0, 0}, Vec4{0, 0, 4, 0}, Vec4{0, 0, 0, 1}}
	if got := s.MulV(Vec4{1, 1, 1, 1}); got != (Vec4{2, 3, 4, 1}) {
		t.Errorf("scale MulV = %v, want (2,3,4,1)", got)
	}
	if MulV(s, v) != s.MulV(v) {
		t.Errorf("MulV free function disagrees with method")
	}
}

func TestScalarBuiltins(t *testing.T) {
	if Clampf(5, 0, 1) != 1 || Clampf(-1, 0, 1) != 0 || Clampf(0.5, 0, 1) != 0.5 {
		t.Errorf("Clampf wrong")
	}
	if Minf(2, 3) != 2 || Maxf(2, 3) != 3 {
		t.Errorf("Minf/Maxf wrong")
	}
	if !almost(Pow(2, 10), 1024) || !almost(Sqrt(16), 4) {
		t.Errorf("Pow/Sqrt wrong")
	}
}
