// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package kernels holds author-once compute kernels: ordinary Go (using gpumath)
// that runs on the CPU, whose source also compiles to MSL/GLSL/SPIR-V via
// poly.red/gpu/shader for the GPU. See specs/foundations/author-once-kernels.md.
package kernels

import . "poly.red/gpu/shader/gpumath"

// Shade is the deferred Blinn-Phong shading kernel, authored once. It runs as Go
// on the CPU (call it per element) and its source (ShadeSrc) compiles to the GPU.
// Inputs are storage buffers; scene = [CamPos.xyz, _, AmbientI, NumLights].
func Shade(gid uint, normals []float32, worldpos []float32, basecol []float32, lights []float32, matidx []float32, materials []float32, scene []float32, out []float32) {
	N := V4(normals[gid*4], normals[gid*4+1], normals[gid*4+2], normals[gid*4+3])
	wpos := V4(worldpos[gid*4], worldpos[gid*4+1], worldpos[gid*4+2], worldpos[gid*4+3])
	col := V4(basecol[gid*4], basecol[gid*4+1], basecol[gid*4+2], basecol[gid*4+3])
	mi := int(matidx[gid])
	diffuse := V4(materials[mi*9], materials[mi*9+1], materials[mi*9+2], materials[mi*9+3])
	specular := V4(materials[mi*9+4], materials[mi*9+5], materials[mi*9+6], materials[mi*9+7])
	shininess := materials[mi*9+8]
	camPos := V4(scene[0], scene[1], scene[2], 0)
	ambientI := scene[4]
	count := int(scene[5])
	acc := col.Scale(ambientI)
	for i := 0; i < count; i++ {
		lt := lights[i*10]
		lp := V4(lights[i*10+1], lights[i*10+2], lights[i*10+3], lights[i*10+4])
		lc := V4(lights[i*10+5], lights[i*10+6], lights[i*10+7], lights[i*10+8])
		li := lights[i*10+9]
		var L Vec4
		var I float32
		if lt < 0.5 {
			Ldir := lp.Sub(wpos)
			L = Normalize(Ldir)
			I = li / Length(Ldir)
		} else {
			L = V4(-lp.X, -lp.Y, -lp.Z, 0)
			I = li
		}
		V := Normalize(camPos.Sub(wpos))
		H := Normalize(L.Add(V))
		Ld := Clampf(Dot(N, L), 0.0, 1.0)
		Ls := Pow(Clampf(Dot(N, H), 0.0, 1.0), shininess)
		acc = acc.Add(diffuse.Mul(col.Scale(Ld * I)).Div(255.0)).Add(specular.Mul(lc.Scale(Ls * I)).Div(255.0))
	}
	out[gid*4] = acc.X
	out[gid*4+1] = acc.Y
	out[gid*4+2] = acc.Z
	out[gid*4+3] = col.W
}
