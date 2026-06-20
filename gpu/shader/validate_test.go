// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package shader

import (
	"strings"
	"testing"
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
		"deferred": deferredKernelSrc,
		"shadow":   shadowKernelSrc,
		"ao":       aoKernelSrc,
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

// The following are verbatim copies of the engine kernels in
// render/gpudeferred.go. Keeping copies here lets the pure-Go compiler tests run
// offline (the render package pulls in cgo-free-but-network-fetched deps). If the
// originals change, update these.

const deferredKernelSrc = `
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

const shadowKernelSrc = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type ShadowU struct {
	W        float32
	DepthLen float32
	N        float32
	Pad      float32
}

func Shadow(gid uint, fragxyz []float32, recv []float32, depths []float32, mats []float32, color []float32, s ShadowU) {
	if recv[gid] < 0.5 {
		return
	}
	fx := fragxyz[gid*4]
	fy := fragxyz[gid*4+1]
	fz := fragxyz[gid*4+2]
	occ := float32(0)
	n := int(s.N)
	dl := int(s.DepthLen)
	width := int(s.W)
	for k := 0; k < n; k++ {
		M := Mat4{
			Vec4{mats[k*16], mats[k*16+1], mats[k*16+2], mats[k*16+3]},
			Vec4{mats[k*16+4], mats[k*16+5], mats[k*16+6], mats[k*16+7]},
			Vec4{mats[k*16+8], mats[k*16+9], mats[k*16+10], mats[k*16+11]},
			Vec4{mats[k*16+12], mats[k*16+13], mats[k*16+14], mats[k*16+15]},
		}
		clip := M * Vec4{fx, fy, fz, 1}
		sx := clip.X / clip.W
		sy := clip.Y / clip.W
		sz := clip.Z / clip.W
		idx := int(sx) + int(sy)*width
		if idx > 0 {
			if idx < dl {
				if sz < depths[k*dl+idx]-0.03 {
					occ = occ + 1
				}
			}
		}
	}
	wf := pow(0.5, occ)
	color[gid*4] = floor(clamp(round(color[gid*4]), 0.0, 255.0) * wf)
	color[gid*4+1] = floor(clamp(round(color[gid*4+1]), 0.0, 255.0) * wf)
	color[gid*4+2] = floor(clamp(round(color[gid*4+2]), 0.0, 255.0) * wf)
}
`

const aoKernelSrc = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

type AOU struct {
	W    float32
	H    float32
	Pad1 float32
	Pad2 float32
}

func AO(gid uint, fragxyz []float32, aoflag []float32, depthbuf []float32, color []float32, s AOU) {
	if aoflag[gid] < 0.5 {
		return
	}
	px := fragxyz[gid*4]
	py := fragxyz[gid*4+1]
	traceDepth := fragxyz[gid*4+2]
	width := int(s.W)
	height := int(s.H)
	total := float32(0)
	for d := 0; d < 8; d++ {
		ang := float32(d) * 0.78539816339744830961
		dirX := cos(ang)
		dirY := sin(ang)
		maxangle := float32(0)
		for t := 0; t < 100; t++ {
			ft := float32(t)
			dx := dirX * ft
			dy := dirY * ft
			distance := sqrt(dx*dx + dy*dy)
			if distance >= 1.0 {
				ix := int(px + dx)
				iy := int(py + dy)
				if ix >= 0 {
					if ix < width {
						if iy >= 0 {
							if iy < height {
								elevation := depthbuf[iy*width+ix] - traceDepth
								maxangle = max(maxangle, atan(elevation/distance))
							}
						}
					}
				}
			}
		}
		total = total + (1.57079632679489661923 - maxangle)
	}
	total = total / (1.57079632679489661923 * 8.0)
	total = pow(total, 10000.0)
	color[gid*4] = floor(clamp(round(color[gid*4]), 0.0, 255.0) * total)
	color[gid*4+1] = floor(clamp(round(color[gid*4+1]), 0.0, 255.0) * total)
	color[gid*4+2] = floor(clamp(round(color[gid*4+2]), 0.0, 255.0) * total)
}
`

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
