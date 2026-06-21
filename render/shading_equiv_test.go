// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image/color"
	"testing"

	"poly.red/buffer"
	"poly.red/geometry/primitive"
	"poly.red/gpu/shader/gpumath/kernels"
	"poly.red/light"
	"poly.red/material"
	"poly.red/math"
	"poly.red/shader"
)

// TestDeferredShadingEquivalence locks the CPU default shading path
// (shader.FragmentShader) to the author-once kernel (kernels.Shade, the same
// source the GPU compiles from kernels.ShadeSrc). They are not merged into one
// function: FragmentShader owns texture/lod, flat-shading and the no-lights
// early return, and remains the fallback for GPU-unsupported scenes. But for the
// shared Blinn-Phong core they must produce the same pixel. This proves "CPU and
// GPU share one shading authority" without a per-fragment slice-kernel rewrite.
//
// The only intrinsic difference is float accumulation order (FragmentShader
// factors Diffuse out of the light sum; the kernel distributes it per light), so
// a 1-LSB tolerance after the shared Round+clamp quantization is expected; a
// larger gap is a real divergence (e.g. the camPos.W / directional seam) to fix,
// not to hide.
func TestDeferredShadingEquivalence(t *testing.T) {
	tex := buffer.NewUniformTexture(color.RGBA{R: 200, G: 150, B: 100, A: 255})
	mat := material.NewBlinnPhong(
		material.Texture(tex),
		material.Diffuse(color.RGBA{R: 220, G: 180, B: 160, A: 255}),
		material.Specular(color.RGBA{R: 255, G: 255, B: 255, A: 255}),
		material.Shininess(32),
	)

	camPos := math.NewVec3[float32](0, 1.5, 3)
	ls := []light.Source{
		light.NewPoint(light.Intensity(3), light.Color(color.RGBA{R: 255, G: 240, B: 220, A: 255}), light.Position(math.NewVec3[float32](-2, 3, 4))),
		light.NewDirectional(light.Intensity(1), light.Color(color.RGBA{R: 180, G: 200, B: 255, A: 255}), light.Direction(math.NewVec3[float32](0, -1, -1))),
	}
	es := []light.Environment{light.NewAmbient(light.Intensity(0.4))}

	// G-buffer fragments spanning many shading angles and positions.
	norms := [][3]float32{
		{0, 1, 0}, {0, 0, 1}, {1, 0, 0}, {0.577, 0.577, 0.577}, {-0.4, 0.8, 0.45}, {0.3, -0.2, 0.93},
	}
	poss := [][3]float32{
		{0, 0, 0}, {1, 0.5, -1}, {-1, 1, 0.5}, {0.2, -0.3, 1}, {-0.6, 0.1, -0.4}, {0.9, 0.9, 0.2},
	}

	// Mirror render/gpudeferred.go's marshaling for the shared inputs.
	materials := []float32{
		float32(mat.Diffuse.R), float32(mat.Diffuse.G), float32(mat.Diffuse.B), float32(mat.Diffuse.A),
		float32(mat.Specular.R), float32(mat.Specular.G), float32(mat.Specular.B), float32(mat.Specular.A),
		mat.Shininess,
	}
	var lightData []float32
	for _, l := range ls {
		switch lt := l.(type) {
		case *light.Point:
			p, c := lt.Position(), lt.Color()
			lightData = append(lightData, 0, p.X, p.Y, p.Z, 1, float32(c.R), float32(c.G), float32(c.B), float32(c.A), lt.Intensity())
		case *light.Directional:
			d, c := lt.Dir(), lt.Color()
			lightData = append(lightData, 1, d.X, d.Y, d.Z, 0, float32(c.R), float32(c.G), float32(c.B), float32(c.A), lt.Intensity())
		}
	}
	var ambientI float32
	for _, e := range es {
		ambientI += e.Intensity()
	}
	scene := []float32{camPos.X, camPos.Y, camPos.Z, 1, ambientI, float32(len(ls)), 0, 0}

	q := func(v float32) uint8 { return uint8(math.Clamp(math.Round(v), 0, 0xff)) }

	for i := range norms {
		nx, ny, nz := norms[i][0], norms[i][1], norms[i][2]
		px, py, pz := poss[i][0], poss[i][1], poss[i][2]
		info := buffer.Fragment{Ok: true, Fragment: primitive.Fragment{
			Nor:     math.NewVec4[float32](nx, ny, nz, 0),
			WordPos: math.NewVec4[float32](px, py, pz, 1),
			Col:     color.RGBA{R: 200, G: 150, B: 100, A: 255},
			U:       0.5, V: 0.5,
		}}

		cpu := shader.FragmentShader(mat, info, camPos, ls, es)

		bc := tex.Query(0, info.U, 1-info.V)
		normals := []float32{nx, ny, nz, 0}
		worldpos := []float32{px, py, pz, 1}
		basecol := []float32{float32(bc.R), float32(bc.G), float32(bc.B), float32(bc.A)}
		out := make([]float32, 4)
		kernels.Shade(0, normals, worldpos, basecol, lightData, []float32{0}, materials, scene, out)

		kr, kg, kb := q(out[0]), q(out[1]), q(out[2])
		if diff(cpu.R, kr) > 1 || diff(cpu.G, kg) > 1 || diff(cpu.B, kb) > 1 {
			t.Errorf("frag %d: FragmentShader=(%d,%d,%d) kernels.Shade=(%d,%d,%d): differ by >1",
				i, cpu.R, cpu.G, cpu.B, kr, kg, kb)
		}
	}
}

func diff(a, b uint8) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}
