// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"

	kernelpkg "poly.red/gpu/shader/gpumath/kernels"
)

// TestCompileRejectsUndefinedIdent is the regression test for the reference
// validation pass: a kernel that reads an identifier which is neither a
// parameter nor a local must fail to compile with a clear "undefined
// identifier" error, instead of silently emitting an undefined name into MSL.
//
// Note on scope: full go/types checking is not used because the kernel DSL
// overloads operators on vector/matrix struct types (e.g. `m * v` where m is a
// Mat4 and v a Vec4), which is not valid Go and which stock go/types rejects.
// This AST-level check resolves bare identifiers against the compiler's own type
// environment instead, catching the common typo/undefined-reference failure.
func TestCompileRejectsUndefinedIdent(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "read of undeclared variable",
			src: `package kernels
func K(gid uint, a []float32, out []float32) {
	out[gid] = a[gid] + missing
}`,
		},
		{
			name: "typo of a parameter",
			src: `package kernels
func K(gid uint, input []float32, out []float32) {
	out[gid] = inputt[gid]
}`,
		},
		{
			name: "use before declaration",
			src: `package kernels
func K(gid uint, out []float32) {
	out[gid] = later
	later := float32(1)
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Compile(tc.src)
			if err == nil {
				t.Fatalf("expected an error, got none")
			}
			if !strings.Contains(err.Error(), "undefined identifier") {
				t.Fatalf("want an undefined-identifier error, got: %v", err)
			}
		})
	}
}

// TestCompileAcceptsRealKernels guards against the validation pass falsely
// rejecting valid kernels. It feeds the actual engine kernels (verbatim copies
// of render/gpudeferred.go's deferred/shadow/AO kernels, which exercise vector
// math, matrix multiply, struct uniforms, loops, var decls, if/else and the full
// builtin set) plus a vertex/fragment pair through Compile and requires success.
// These cannot be run via the darwin GPU tests on a non-darwin host, so this
// pure-Go test is the offline regression guard for the corpus.
func TestCompileAcceptsRealKernels(t *testing.T) {
	corpus := map[string]string{
		"deferred": kernelpkg.ShadeSrc,
		"shadow":   kernelpkg.ShadowSrc,
		"ao":       kernelpkg.AOSrc,
		"vertfrag": vertFragKernelSrc,
	}
	for name, src := range corpus {
		t.Run(name, func(t *testing.T) {
			ks, err := Compile(src)
			if err != nil {
				t.Fatalf("compile %s: %v", name, err)
			}
			if len(ks) == 0 {
				t.Fatalf("compile %s: no kernels produced", name)
			}
		})
	}
}

// The deferred, shadow, and ao kernels come from the canonical author-once
// sources (kernels.ShadeSrc / ShadowSrc / AOSrc) so this corpus cannot drift
// from the engine. Only vertfrag below is a local copy: an example
// vertex/fragment pair that is not a kernels-package kernel. Keeping it here lets
// the pure-Go compiler tests run offline.

const vertFragKernelSrc = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type VOut struct {
	Pos   Vec4 ` + "`gpu:\"position\"`" + `
	Color Vec4
}

//gpu:vertex
func VMain(vid uint, pos []float32, col []float32) VOut {
	return VOut{Vec4{pos[vid*2], pos[vid*2+1], 0, 1}, Vec4{col[vid*3], col[vid*3+1], col[vid*3+2], 1}}
}

//gpu:fragment
func FMain(in VOut) Vec4 {
	return in.Color
}
`
