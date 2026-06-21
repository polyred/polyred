// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"
)

// TestGpumathLowering verifies the author-once mechanism: a kernel written in
// gpumath method/free-function form (valid Go that runs on the CPU) lowers to the
// same shader source as the equivalent operator-DSL kernel. So one Go source can
// drive both the CPU (as Go) and the GPU (compiled). See
// specs/foundations/author-once-kernels.md.
func TestGpumathLowering(t *testing.T) {
	// gpumath method form (this is valid Go when gpumath is dot-imported).
	gm := `package k
func F(gid uint, a []float32, b []float32, out []float32) {
	u := Vec4{a[gid*4], a[gid*4+1], a[gid*4+2], a[gid*4+3]}
	v := Vec4{b[gid*4], b[gid*4+1], b[gid*4+2], b[gid*4+3]}
	r := Normalize(u.Sub(v))
	s := Clampf(Dot(u, v), 0.0, 1.0)
	w := u.Scale(s).Add(v.Div(255.0))
	out[gid*4] = r.X + w.Y
}`
	for _, tc := range []struct {
		name    string
		compile func(string) (map[string]*Kernel, error)
		get     func(*Kernel) string
		want    []string
	}{
		{"MSL", Compile, func(k *Kernel) string { return k.MSL }, []string{
			"normalize((u - v))", "clamp(dot(u, v), 0.0, 1.0)", "(u * s)", "(v / 255.0)",
		}},
		{"GLSL", CompileGLSL, func(k *Kernel) string { return k.GLSL }, []string{
			"normalize((u - v))", "clamp(dot(u, v), 0.0, 1.0)", "(u * s)", "(v / 255.0)",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ks, err := tc.compile(gm)
			if err != nil {
				t.Fatalf("compile: %v", err)
			}
			src := tc.get(ks["F"])
			for _, w := range tc.want {
				if !strings.Contains(src, w) {
					t.Errorf("lowered %s missing %q:\n%s", tc.name, w, src)
				}
			}
		})
	}
}
