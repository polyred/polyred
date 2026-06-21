// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Cross-backend parity harness. A single Go shading kernel (Blinn-Phong over a
// synthetic G-buffer, storage buffers only so it runs unchanged on every
// backend) is compiled per-backend, run through the public Device API, and
// compared to one shared Go reference (the CPU oracle). The platform entry
// points (parity_darwin_test.go: Metal; parity_linux_test.go: GL + Vulkan) call
// runShadingParity with a backend-specific module builder. Because every backend
// is checked against the same oracle, passing on each CI job proves the backends
// agree with each other (and the CPU) within tolerance.
package gpu_test

import (
	"math"
	"testing"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader"
)

func parityBytes(d []float32) []byte {
	if len(d) == 0 {
		return nil
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}

func parityFloats(b []byte, n int) []float32 {
	out := make([]float32, n)
	copy(out, unsafe.Slice((*float32)(unsafe.Pointer(&b[0])), n))
	return out
}

// shadingKernelSrc is a storage-buffer-only Blinn-Phong deferred shading kernel
// (the engine's deferred kernel with its Scene uniform passed as a buffer, so no
// UBO descriptor differences between backends).
const shadingKernelSrc = `
package kernels

type Vec4 struct{ X, Y, Z, W float32 }

func Shade(gid uint, normals []float32, worldpos []float32, basecol []float32, lights []float32, matidx []float32, materials []float32, scene []float32, out []float32) {
	N := Vec4{normals[gid*4], normals[gid*4+1], normals[gid*4+2], normals[gid*4+3]}
	wpos := Vec4{worldpos[gid*4], worldpos[gid*4+1], worldpos[gid*4+2], worldpos[gid*4+3]}
	col := Vec4{basecol[gid*4], basecol[gid*4+1], basecol[gid*4+2], basecol[gid*4+3]}
	mi := int(matidx[gid])
	diffuse := Vec4{materials[mi*9], materials[mi*9+1], materials[mi*9+2], materials[mi*9+3]}
	specular := Vec4{materials[mi*9+4], materials[mi*9+5], materials[mi*9+6], materials[mi*9+7]}
	shininess := materials[mi*9+8]
	camPos := Vec4{scene[0], scene[1], scene[2], 0}
	ambientI := scene[4]
	count := int(scene[5])
	acc := col * ambientI
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
		V := normalize(camPos - wpos)
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

type pv4 struct{ x, y, z, w float32 }

func (a pv4) sub(b pv4) pv4       { return pv4{a.x - b.x, a.y - b.y, a.z - b.z, a.w - b.w} }
func (a pv4) add(b pv4) pv4       { return pv4{a.x + b.x, a.y + b.y, a.z + b.z, a.w + b.w} }
func (a pv4) mul(b pv4) pv4       { return pv4{a.x * b.x, a.y * b.y, a.z * b.z, a.w * b.w} }
func (a pv4) scale(s float32) pv4 { return pv4{a.x * s, a.y * s, a.z * s, a.w * s} }
func (a pv4) dot(b pv4) float32   { return a.x*b.x + a.y*b.y + a.z*b.z + a.w*b.w }
func (a pv4) length() float32     { return float32(math.Sqrt(float64(a.dot(a)))) }
func (a pv4) normalize() pv4 {
	l := a.length()
	if l == 0 {
		return a
	}
	return a.scale(1 / l)
}

func pclamp(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// parityScene is the shared synthetic G-buffer + lights + materials.
type parityScene struct {
	n         int
	normals   []float32
	worldpos  []float32
	basecol   []float32
	lights    []float32
	matidx    []float32
	materials []float32
	scene     []float32
}

func makeParityScene(n int) parityScene {
	s := parityScene{n: n}
	for i := 0; i < n; i++ {
		f := float32(i)
		// A varied but deterministic normal (normalized), position and colour.
		nrm := pv4{float32(math.Sin(float64(f) * 0.3)), 0.5, float32(math.Cos(float64(f) * 0.2)), 0}.normalize()
		s.normals = append(s.normals, nrm.x, nrm.y, nrm.z, 0)
		s.worldpos = append(s.worldpos, f*0.01, 0.2, -float32(i%5)*0.1, 0)
		s.basecol = append(s.basecol, 40+float32(i%200), 80+float32(i%150), 120+float32(i%100), 255)
		s.matidx = append(s.matidx, float32(i%2))
	}
	// Two materials (diffuse rgba, specular rgba, shininess).
	s.materials = []float32{
		1, 1, 1, 1, 0.8, 0.8, 0.8, 1, 32,
		0.9, 0.7, 0.5, 1, 1, 1, 1, 1, 8,
	}
	// Two lights: one point, one directional.
	s.lights = []float32{
		0, 2, 3, 4, 0, 1, 1, 1, 0, 3, // point: lt=0, lp, lc, li
		1, -1, -1, -1, 0, 0.6, 0.6, 0.7, 0, 1, // directional: lt=1, dir, lc, li
	}
	// scene: CamPos.xyz, _, AmbientI, NumLights.
	s.scene = []float32{0, 0.6, 0.9, 0, 0.4, 2}
	return s
}

func vec4At(b []float32, i int) pv4 { return pv4{b[i*4], b[i*4+1], b[i*4+2], b[i*4+3]} }

// cpuShade is the CPU oracle: the Blinn-Phong shading replicated in Go (float32),
// matching shadingKernelSrc.
func cpuShade(s parityScene) []float32 {
	out := make([]float32, s.n*4)
	for g := 0; g < s.n; g++ {
		N := vec4At(s.normals, g)
		wpos := vec4At(s.worldpos, g)
		col := vec4At(s.basecol, g)
		mi := int(s.matidx[g])
		diff := pv4{s.materials[mi*9], s.materials[mi*9+1], s.materials[mi*9+2], s.materials[mi*9+3]}
		spec := pv4{s.materials[mi*9+4], s.materials[mi*9+5], s.materials[mi*9+6], s.materials[mi*9+7]}
		shin := s.materials[mi*9+8]
		camPos := pv4{s.scene[0], s.scene[1], s.scene[2], 0}
		ambientI := s.scene[4]
		count := int(s.scene[5])
		acc := col.scale(ambientI)
		for i := 0; i < count; i++ {
			lt := s.lights[i*10]
			lp := pv4{s.lights[i*10+1], s.lights[i*10+2], s.lights[i*10+3], s.lights[i*10+4]}
			lc := pv4{s.lights[i*10+5], s.lights[i*10+6], s.lights[i*10+7], s.lights[i*10+8]}
			li := s.lights[i*10+9]
			var L pv4
			var I float32
			if lt < 0.5 {
				Ldir := lp.sub(wpos)
				L = Ldir.normalize()
				I = li / Ldir.length()
			} else {
				L = pv4{-lp.x, -lp.y, -lp.z, 0}
				I = li
			}
			V := camPos.sub(wpos).normalize()
			H := L.add(V).normalize()
			Ld := pclamp(N.dot(L), 0, 1)
			Ls := float32(math.Pow(float64(pclamp(N.dot(H), 0, 1)), float64(shin)))
			acc = acc.add(diff.mul(col.scale(Ld * I)).scale(1.0 / 255.0)).add(spec.mul(lc.scale(Ls * I)).scale(1.0 / 255.0))
		}
		out[g*4] = acc.x
		out[g*4+1] = acc.y
		out[g*4+2] = acc.z
		out[g*4+3] = col.w
	}
	return out
}

// runShadingParity runs the shared shading kernel on dev via mk (a backend
// specific module + bindings builder) and compares the result to the CPU oracle.
func runShadingParity(t *testing.T, dev *gpu.Device, mk func(goSrc, entry string) (*gpu.ShaderModule, []shader.Binding, error)) {
	t.Helper()
	const n = 64
	sc := makeParityScene(n)
	want := cpuShade(sc)

	mod, binds, err := mk(shadingKernelSrc, "Shade")
	if err != nil {
		t.Fatalf("compile kernel for %v: %v", dev.Driver(), err)
	}

	inputs := map[string][]float32{
		"normals": sc.normals, "worldpos": sc.worldpos, "basecol": sc.basecol,
		"lights": sc.lights, "matidx": sc.matidx, "materials": sc.materials,
		"scene": sc.scene, "out": make([]float32, n*4),
	}

	var le []gpu.BindGroupLayoutEntry
	var bge []gpu.BindGroupEntry
	bufByName := map[string]*gpu.Buffer{}
	for _, b := range binds {
		data, ok := inputs[b.Name]
		if !ok {
			t.Fatalf("kernel binding %q has no input", b.Name)
		}
		buf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: parityBytes(data), Usage: gpu.BufferStorage})
		if err != nil {
			t.Fatalf("buffer %q: %v", b.Name, err)
		}
		bufByName[b.Name] = buf
		le = append(le, gpu.BindGroupLayoutEntry{Binding: b.Index, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer})
		bge = append(bge, gpu.BindGroupEntry{Binding: b.Index, Buffer: buf})
	}
	layout := dev.NewBindGroupLayout(le...)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "Shade",
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	bg := dev.NewBindGroup(layout, bge...)

	enc := dev.NewCommandEncoder()
	pass := enc.BeginComputePass()
	pass.SetPipeline(pipe)
	pass.SetBindGroup(0, bg)
	pass.Dispatch(n, 1, 1)
	pass.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	got := parityFloats(bufByName["out"].Bytes(), n*4)

	// pow/normalize/sqrt differ slightly across drivers; assert closeness.
	const tol = 0.05
	var maxDiff float64
	for i := range want {
		d := math.Abs(float64(got[i] - want[i]))
		if d > maxDiff {
			maxDiff = d
		}
		if d > tol {
			t.Fatalf("%v parity: out[%d]=%v want %v (diff %v > tol %v)", dev.Driver(), i, got[i], want[i], d, tol)
		}
	}
	t.Logf("%v shading parity: %d pixels match the CPU oracle (max diff %.6f)", dev.Driver(), n, maxDiff)
}
