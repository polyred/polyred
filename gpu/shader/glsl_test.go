// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"
)

// TestCompileGLSLCompute checks the GLSL ES 3.10 compute emitter on the real
// engine kernels: the right preamble/layout decls are produced, MSL spellings do
// not leak (float4 -> vec4), the reserved word `out` is mangled, matrix multiply
// and uint arithmetic survive, and the binding metadata uses separate SSBO/UBO
// index spaces. These are structural (text) checks; executing the shaders needs
// a Linux GLES 3.1 device and is gated there (see specs/foundations/
// gpu-gl-backend.md).
func TestCompileGLSLCompute(t *testing.T) {
	ks, err := CompileGLSL(deferredKernelSrc)
	if err != nil {
		t.Fatalf("compile deferred: %v", err)
	}
	g := ks["Shade"].GLSL

	for _, want := range []string{
		"#version 310 es",
		"layout(local_size_x = 1) in;",
		"layout(std430, binding = 0) readonly buffer _ssbo0 { float normals[]; };",
		"layout(std140, binding = 0) uniform _ubo0 {",
		"int gid = int(gl_GlobalInvocationID.x);",
		"void main()",
	} {
		if !strings.Contains(g, want) {
			t.Errorf("deferred GLSL missing %q\n---\n%s", want, g)
		}
	}
	// MSL spellings must not leak into GLSL.
	if strings.Contains(g, "float4") || strings.Contains(g, "float2") {
		t.Errorf("deferred GLSL leaked MSL vector spelling:\n%s", g)
	}
	// `out` is a GLSL keyword; the output buffer must be mangled, in both its
	// declaration and its uses.
	if strings.Contains(g, " out[") || strings.Contains(g, "float out[]") {
		t.Errorf("deferred GLSL did not mangle reserved word 'out':\n%s", g)
	}
	if !strings.Contains(g, "out_[") {
		t.Errorf("deferred GLSL missing mangled out_:\n%s", g)
	}

	// Binding metadata: storage buffers and the uniform block use separate index
	// spaces (the uniform is binding 0 in the UBO space, not 6).
	var sawUBO0 bool
	for _, b := range ks["Shade"].Bindings {
		if b.Kind == UniformBuffer && b.Index == 0 {
			sawUBO0 = true
		}
	}
	if !sawUBO0 {
		t.Errorf("expected the uniform block at UBO binding 0, got %+v", ks["Shade"].Bindings)
	}

	// Matrix multiply (shadow kernel) lowers to mat4 * vec4.
	sk, err := CompileGLSL(shadowKernelSrc)
	if err != nil {
		t.Fatalf("compile shadow: %v", err)
	}
	if !strings.Contains(sk["Shadow"].GLSL, "mat4") {
		t.Errorf("shadow GLSL missing mat4:\n%s", sk["Shadow"].GLSL)
	}

	// AO kernel uses trig/atan and nested loops; just require it compiles.
	if _, err := CompileGLSL(aoKernelSrc); err != nil {
		t.Fatalf("compile AO: %v", err)
	}
}

// TestCompileGLSLRejectsUnsupported verifies the GLSL compute emitter rejects
// what it does not yet support, with a clear error, rather than emitting bad
// shader source.
func TestCompileGLSLRejectsUnsupported(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		{
			name: "vertex stage",
			src: `package k
type Vec4 struct{ X, Y, Z, W float32 }
//gpu:vertex
func V(vid uint, pos []float32) Vec4 { return Vec4{pos[vid], 0, 0, 1} }`,
		},
		{
			name: "texture param",
			src: `package k
type Vec4 struct{ X, Y, Z, W float32 }
func K(gid uint, tex Texture2D, samp Sampler, out []float32) {
	c := tex.Sample(samp, Vec2{0, 0})
	out[gid] = c.X
}`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := CompileGLSL(tc.src); err == nil {
				t.Fatalf("expected an error, got none")
			}
		})
	}
}
