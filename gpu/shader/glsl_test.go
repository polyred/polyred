// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"

	kernelpkg "poly.red/gpu/shader/gpumath/kernels"
)

// uniformSceneKernelSrc is a synthetic Blinn-Phong kernel that takes its scene
// data as a std140 uniform struct (the `s Scene` param) alongside storage
// buffers. It is a compiler fixture for the dual SSBO/UBO index-space path, NOT
// the engine's kernel: the engine's real deferred shader is the author-once
// kernels.Shade, which passes the scene as a storage buffer (no uniform). This
// fixture keeps the uniform-struct param on purpose so the UBO emitter stays
// covered.
const uniformSceneKernelSrc = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type Scene struct {
	CamPos    Vec4
	AmbientI  float32
	NumLights float32
	Pad1      float32
	Pad2      float32
}

func Shade(gid uint, normals []float32, worldpos []float32, basecol []float32, lights []float32, matidx []float32, materials []float32, s Scene, out []float32) {
	N := Vec4{normals[gid*4], normals[gid*4+1], normals[gid*4+2], normals[gid*4+3]}
	wpos := Vec4{worldpos[gid*4], worldpos[gid*4+1], worldpos[gid*4+2], worldpos[gid*4+3]}
	col := Vec4{basecol[gid*4], basecol[gid*4+1], basecol[gid*4+2], basecol[gid*4+3]}

	mi := int(matidx[gid])
	diffuse := Vec4{materials[mi*9], materials[mi*9+1], materials[mi*9+2], materials[mi*9+3]}
	specular := Vec4{materials[mi*9+4], materials[mi*9+5], materials[mi*9+6], materials[mi*9+7]}
	shininess := materials[mi*9+8]

	acc := col * s.AmbientI
	count := int(s.NumLights)
	for i := 0; i < count; i++ {
		lt := lights[i*10]
		lp := Vec4{lights[i*10+1], lights[i*10+2], lights[i*10+3], lights[i*10+4]}
		lc := Vec4{lights[i*10+5], lights[i*10+6], lights[i*10+7], lights[i*10+8]}
		li := lights[i*10+9]
		var L Vec4
		var I float32
		if lt < 0.5 {
			Ldir := lp - wpos
			L = normalize(Ldir)
			I = li / length(Ldir)
		} else {
			L = Vec4{-lp.X, -lp.Y, -lp.Z, 0}
			I = li
		}
		V := normalize(s.CamPos - wpos)
		H := normalize(L + V)
		Ld := clamp(dot(N, L), 0.0, 1.0)
		Ls := pow(clamp(dot(N, H), 0.0, 1.0), shininess)
		acc = acc + diffuse*(col*(Ld*I))/255.0 + specular*(lc*(Ls*I))/255.0
	}
	out[gid*4] = acc.X
	out[gid*4+1] = acc.Y
	out[gid*4+2] = acc.Z
	out[gid*4+3] = col.W
}
`

// TestCompileGLSLCompute checks the GLSL ES 3.10 compute emitter on a kernel that
// mixes storage and uniform buffers: the right preamble/layout decls are produced,
// MSL spellings do not leak (float4 -> vec4), the reserved word `out` is mangled,
// matrix multiply and uint arithmetic survive, and the binding metadata uses
// separate SSBO/UBO index spaces. These are structural (text) checks; executing
// the shaders needs a Linux GLES 3.1 device and is gated there (see specs/
// foundations/gpu-gl-backend.md).
func TestCompileGLSLCompute(t *testing.T) {
	ks, err := CompileGLSL(uniformSceneKernelSrc)
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
	sk, err := CompileGLSL(kernelpkg.ShadowSrc)
	if err != nil {
		t.Fatalf("compile shadow: %v", err)
	}
	if !strings.Contains(sk["Shadow"].GLSL, "mat4") {
		t.Errorf("shadow GLSL missing mat4:\n%s", sk["Shadow"].GLSL)
	}

	// AO kernel uses trig/atan and nested loops; just require it compiles.
	if _, err := CompileGLSL(kernelpkg.AOSrc); err != nil {
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
