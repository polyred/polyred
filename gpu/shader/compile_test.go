// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"
)

// kernels is the Go source for the matrix compute kernels. The compiler turns
// these into MSL equivalent to the hand-written gpu/tests/shaders/math.metal.
const kernels = `
package kernels

type Params struct {
	WidthA uint
	HeightA uint
	WidthB uint
}

func Add(gid int, a []float32, b []float32, out []float32) {
	out[gid] = a[gid] + b[gid]
}

func Sub(gid int, a []float32, b []float32, out []float32) {
	out[gid] = a[gid] - b[gid]
}

func Sqrt(gid int, a []float32, out []float32) {
	out[gid] = sqrt(a[gid])
}

func Mul(gid int, a []float32, b []float32, out []float32, p Params) {
	row := uint(gid) / p.WidthB
	col := uint(gid) % p.WidthB
	var sum float32 = 0
	for i := uint(0); i < p.WidthA; i++ {
		sum += a[row*p.WidthA+i] * b[i*p.WidthB+col]
	}
	out[gid] = sum
}
`

func TestCompile(t *testing.T) {
	ks, err := Compile(kernels)
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	for _, name := range []string{"Add", "Sub", "Sqrt", "Mul"} {
		if _, ok := ks[name]; !ok {
			t.Fatalf("kernel %q not compiled", name)
		}
	}

	add := ks["Add"]
	if len(add.Bindings) != 3 {
		t.Fatalf("Add: want 3 bindings, got %d", len(add.Bindings))
	}
	// a and b are read-only (const), out is written (non-const).
	for _, want := range []string{
		"device const float* a [[buffer(0)]]",
		"device const float* b [[buffer(1)]]",
		"device float* out [[buffer(2)]]",
		"uint gid [[thread_position_in_grid]]",
		"out[gid] = (a[gid] + b[gid]);",
	} {
		if !strings.Contains(add.MSL, want) {
			t.Fatalf("Add MSL missing %q\n---\n%s", want, add.MSL)
		}
	}

	mul := ks["Mul"]
	for _, want := range []string{
		"struct Params {",
		"constant Params& p [[buffer(3)]]",
		"uint row = (uint(gid) / p.WidthB);",
		"float sum = 0;",
		"for (uint i = uint(0); (i < p.WidthA); i++) {",
		"sum += (a[((row * p.WidthA) + i)] * b[((i * p.WidthB) + col)]);",
	} {
		if !strings.Contains(mul.MSL, want) {
			t.Fatalf("Mul MSL missing %q\n---\n%s", want, mul.MSL)
		}
	}
}

func TestCompileRejectsUnsupported(t *testing.T) {
	bad := `package k
func K(gid int, a []float32) {
	go func() {}()
}`
	if _, err := Compile(bad); err == nil {
		t.Fatal("expected error for goroutine in kernel, got nil")
	}
}
