// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// GPU forward-rasterizer measurement + bring-up (brick 3b, gpu-forward-raster.md).
// Two CI checks on Mesa llvmpipe (surfaceless):
//
//   - TestGPUForwardRasterCoverage (step 1): the RASTER only. Screen-space
//     vertices are computed on the CPU exactly as draw() (so the transform is
//     identical) and rasterized on the GPU; coverage is compared to the CPU
//     forward pass. Result so far: pixel-identical.
//   - TestGPUForwardTransformCoverage (step 2a): the GPU does the VERTEX
//     TRANSFORM too. Model-space positions + the per-object trans = Proj*View*Model
//     matrix are uploaded; the vertex shader computes gl_Position = -(trans*pos)
//     (the negation matches the renderer's projection, whose w is negated -- the
//     CPU divides by +w via Pos() while clip.w<0), and glViewport reproduces the
//     renderer's ViewportMatrix exactly. Coverage is compared to the CPU forward
//     pass, validating the GPU vertex pipeline end to end.
package render

import (
	stdmath "math"
	"os"
	"testing"

	"poly.red/camera"
	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/gpu"
	"poly.red/math"
	"poly.red/scene"
)

// cpuForwardCoverage runs the CPU forward pass for the scene and returns a
// per-pixel coverage mask (a fragment was written; depth clears to 0). It calls
// passForward directly because Render() defers NextBuffer(), which clears and
// rotates to an empty buffer.
func cpuForwardCoverage(s *scene.Scene, c camera.Interface, w, h int) (cov []bool, n int) {
	r := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU())
	buf := r.CurrBuffer()
	buf.Clear()
	r.passForward()
	depth := buf.Depth()
	cov = make([]bool, w*h)
	for i := 0; i < w*h; i++ {
		if depth.Pix[i*4] > 0 {
			cov[i] = true
			n++
		}
	}
	return cov, n
}

func TestGPUForwardRasterCoverage(t *testing.T) {
	dev := openGLOrSkip(t)
	defer dev.Close()

	const w, h = 128, 128
	s, c := newscene(w, h)
	cpuCov, cpuN := cpuForwardCoverage(s, c, w, h)

	// Screen-space vertices computed on the CPU, mapped to NDC with w=1 so the GPU
	// does no perspective divide.
	view, proj := c.ViewMatrix(), c.ProjMatrix()
	viewport := math.ViewportMatrix(float32(w), float32(h))
	var ndc []float32
	scene.IterObjects(s, func(g *geometry.Geometry, model math.Mat4[float32]) bool {
		trans := proj.MulM(view).MulM(model.MulM(g.ModelMatrix()))
		for _, tri := range g.Triangles() {
			for _, v := range []*primitive.Vertex{tri.V1, tri.V2, tri.V3} {
				sp := trans.MulV(v.Pos).Apply(viewport).Pos()
				ndc = append(ndc, 2*sp.X/float32(w)-1, 2*sp.Y/float32(h)-1, 0, 1)
			}
		}
		return true
	})
	gpuCov := gpuRasterCoverage(t, dev, ndc, w, h)
	reportCoverage(t, "raster (CPU transform)", cpuCov, gpuCov, cpuN, w, h)
}

func TestGPUForwardTransformCoverage(t *testing.T) {
	dev := openGLOrSkip(t)
	defer dev.Close()

	const w, h = 128, 128
	s, c := newscene(w, h)
	cpuCov, cpuN := cpuForwardCoverage(s, c, w, h)

	// Upload model-space positions + the per-object trans matrix (column-major for
	// GLSL) and let the GPU do the transform.
	view, proj := c.ViewMatrix(), c.ProjMatrix()
	var objs []transObject
	scene.IterObjects(s, func(g *geometry.Geometry, model math.Mat4[float32]) bool {
		trans := proj.MulM(view).MulM(model.MulM(g.ModelMatrix()))
		o := transObject{mat: colMajor(trans)}
		for _, tri := range g.Triangles() {
			for _, v := range []*primitive.Vertex{tri.V1, tri.V2, tri.V3} {
				o.pos = append(o.pos, v.Pos.X, v.Pos.Y, v.Pos.Z, v.Pos.W)
			}
		}
		objs = append(objs, o)
		return true
	})
	gpuCov := gpuTransformCoverage(t, dev, objs, w, h)
	reportCoverage(t, "raster+transform (GPU)", cpuCov, gpuCov, cpuN, w, h)
}

type transObject struct {
	pos []float32  // model-space positions, 4 floats/vertex
	mat [16]float32 // Proj*View*Model, column-major
}

// reportCoverage logs the CPU/GPU coverage delta (the deliverable) and guards
// against a gross mapping error; it is flip-robust on the framebuffer Y origin.
func reportCoverage(t *testing.T, label string, cpuCov, gpuCov []bool, cpuN, w, h int) {
	t.Helper()
	gpuN := 0
	for _, ok := range gpuCov {
		if ok {
			gpuN++
		}
	}
	diff := coverageDiff(cpuCov, gpuCov, w, h, false)
	diffFlip := coverageDiff(cpuCov, gpuCov, w, h, true)
	flipped := diffFlip < diff
	if flipped {
		diff = diffFlip
	}
	frac := float64(diff) / float64(w*h)
	t.Logf("forward %s: cpu=%d gpu=%d differ=%d (%.2f%% of %d px, yflip=%v)",
		label, cpuN, gpuN, diff, frac*100, w*h, flipped)
	if cpuN == 0 || gpuN == 0 {
		t.Fatalf("one side rendered nothing (cpu=%d gpu=%d)", cpuN, gpuN)
	}
	if frac > 0.25 {
		t.Fatalf("CPU/GPU forward coverage differs on %.2f%% of pixels (gross mismatch; see log)", frac*100)
	}
}

func coverageDiff(a, b []bool, w, h int, flipY bool) int {
	n := 0
	for y := 0; y < h; y++ {
		by := y
		if flipY {
			by = h - 1 - y
		}
		for x := 0; x < w; x++ {
			if a[y*w+x] != b[by*w+x] {
				n++
			}
		}
	}
	return n
}

func openGLOrSkip(t *testing.T) *gpu.Device {
	t.Helper()
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GPU forward raster tests")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	return dev
}

// colMajor flattens a row-major Mat4 into the column-major order GLSL's mat4(...)
// constructor expects, so GLSL `mat4(arr) * p` equals the renderer's m.MulV(p).
func colMajor(m math.Mat4[float32]) [16]float32 {
	var a [16]float32
	for col := 0; col < 4; col++ {
		for row := 0; row < 4; row++ {
			a[col*4+row] = m.Get(row, col)
		}
	}
	return a
}

const fwdCovFrag = `#version 310 es
precision highp float;
out vec4 fragColor;
void main() { fragColor = vec4(1.0, 1.0, 1.0, 1.0); }`

const fwdRasterVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _v { float pos[]; };
void main() {
	int i = gl_VertexID;
	gl_Position = vec4(pos[i*4], pos[i*4+1], pos[i*4+2], pos[i*4+3]);
}`

const fwdTransformVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _pos { float pos[]; };
layout(std430, binding = 1) readonly buffer _mat { float m[]; };
void main() {
	int i = gl_VertexID;
	vec4 p = vec4(pos[i*4], pos[i*4+1], pos[i*4+2], pos[i*4+3]);
	mat4 M = mat4(m[0], m[1], m[2], m[3], m[4], m[5], m[6], m[7],
	              m[8], m[9], m[10], m[11], m[12], m[13], m[14], m[15]);
	// The renderer's projection yields a negated w (the CPU divides by +w via
	// Pos() with clip.w<0); negating the whole vector makes the GPU's divide-by-w
	// reproduce the same NDC, and glViewport matches ViewportMatrix exactly.
	gl_Position = -(M * p);
}`

// gpuRasterCoverage rasterizes NDC triangles (4 floats/vertex) and returns a
// coverage mask. No depth: silhouette coverage is depth-independent.
func gpuRasterCoverage(t *testing.T, dev *gpu.Device, ndc []float32, w, h int) []bool {
	t.Helper()
	if len(ndc) == 0 {
		t.Fatal("no triangles")
	}
	pipe := buildPipe(t, dev, fwdRasterVert)
	color := mkColor(t, dev, w, h)
	vbuf := mkBuf(t, dev, ndc)
	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: color, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, len(ndc)/4)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	return coverageOf(color.ReadPixels(), w, h)
}

// gpuTransformCoverage rasterizes objects whose model-space positions are
// transformed on the GPU by each object's trans matrix, with depth testing so
// overlapping objects compose correctly.
func gpuTransformCoverage(t *testing.T, dev *gpu.Device, objs []transObject, w, h int) []bool {
	t.Helper()
	if len(objs) == 0 {
		t.Fatal("no objects")
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mkMod(t, dev, fwdTransformVert), VertexEntry: "main",
		FragmentModule: mkMod(t, dev, fwdCovFrag), FragmentEntry: "main",
		ColorFormat: gpu.RGBA8Unorm,
		DepthFormat: gpu.Depth32Float,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	color := mkColor(t, dev, w, h)
	depth, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.Depth32Float, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		t.Fatalf("depth texture: %v", err)
	}
	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: color, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
		DepthTexture: depth, ClearDepth: 1,
	})
	rp.SetPipeline(pipe)
	for _, o := range objs {
		rp.SetVertexBuffer(0, mkBuf(t, dev, o.pos))
		rp.SetVertexBuffer(1, mkBuf(t, dev, o.mat[:]))
		rp.Draw(gpu.TriangleList, 0, len(o.pos)/4)
	}
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	return coverageOf(color.ReadPixels(), w, h)
}

func buildPipe(t *testing.T, dev *gpu.Device, vert string) *gpu.RenderPipeline {
	t.Helper()
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mkMod(t, dev, vert), VertexEntry: "main",
		FragmentModule: mkMod(t, dev, fwdCovFrag), FragmentEntry: "main",
		ColorFormat: gpu.RGBA8Unorm,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	return pipe
}

func mkMod(t *testing.T, dev *gpu.Device, src string) *gpu.ShaderModule {
	t.Helper()
	m, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: src})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	return m
}

func mkColor(t *testing.T, dev *gpu.Device, w, h int) *gpu.Texture {
	t.Helper()
	tex, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		t.Fatalf("color texture: %v", err)
	}
	return tex
}

func mkBuf(t *testing.T, dev *gpu.Device, d []float32) *gpu.Buffer {
	t.Helper()
	b, err := dev.NewBuffer(gpu.BufferDescriptor{Data: floatsToBytes(d), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("buffer: %v", err)
	}
	return b
}

func coverageOf(pix []byte, w, h int) []bool {
	cov := make([]bool, w*h)
	for i := 0; i < w*h; i++ {
		cov[i] = pix[i*4] > 127
	}
	return cov
}

func floatsToBytes(d []float32) []byte {
	b := make([]byte, len(d)*4)
	for i, f := range d {
		u := stdmath.Float32bits(f)
		b[i*4], b[i*4+1], b[i*4+2], b[i*4+3] = byte(u), byte(u>>8), byte(u>>16), byte(u>>24)
	}
	return b
}
