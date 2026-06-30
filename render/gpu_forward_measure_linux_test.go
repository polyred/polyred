// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build linux

// GPU forward-rasterizer measurement (brick 3b, step 1, gpu-forward-raster.md).
// Before designing a parity gate we must MEASURE how a GPU hardware rasterizer
// disagrees with the CPU scanline rasterizer. This isolates the raster itself:
// the per-triangle clip-space vertices are computed on the CPU (identical to
// passForward's transform), so only coverage/edge/depth handling differs on the
// GPU. It compares the silhouette (which pixels are covered) of the CPU forward
// pass against a GPU depth-tested raster of the same triangles, logs the delta
// distribution, and asserts only a deliberately loose bound for now -- the gate
// is tightened to the measured number once CI shows it. Runs on Mesa llvmpipe
// (surfaceless).
package render

import (
	stdmath "math"
	"os"
	"testing"

	"poly.red/geometry"
	"poly.red/geometry/primitive"
	"poly.red/gpu"
	"poly.red/math"
	"poly.red/scene"
)

func TestGPUForwardRasterCoverage(t *testing.T) {
	if os.Getenv("EGL_PLATFORM") != "surfaceless" {
		t.Skip("set EGL_PLATFORM=surfaceless to run the GPU forward raster measurement")
	}
	dev, err := gpu.Open(gpu.WithDriver(gpu.DriverGL))
	if err != nil {
		t.Skipf("no GL device: %v", err)
	}
	defer dev.Close()

	const w, h = 128, 128
	s, c := newscene(w, h)

	// CPU forward pass: coverage = any pixel that received a fragment (depth left
	// at the zero clear value where nothing was drawn). MSAA(1) so the CPU buffer
	// is at output resolution, matching the GPU's single-sample raster.
	r := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU())
	r.Render()
	depth := r.CurrBuffer().Depth()
	cpuCov := make([]bool, w*h)
	cpuN := 0
	for i := 0; i < w*h; i++ {
		if depth.Pix[i*4] > 0 {
			cpuCov[i] = true
			cpuN++
		}
	}

	// GPU raster of the SAME triangles in clip space (transform done on the CPU so
	// only the rasterization differs).
	view, proj := c.ViewMatrix(), c.ProjMatrix()
	var clip []float32
	scene.IterObjects(s, func(g *geometry.Geometry, model math.Mat4[float32]) bool {
		trans := proj.MulM(view).MulM(model.MulM(g.ModelMatrix()))
		for _, tri := range g.Triangles() {
			for _, v := range []*primitive.Vertex{tri.V1, tri.V2, tri.V3} {
				p := trans.MulV(v.Pos)
				clip = append(clip, p.X, p.Y, p.Z, p.W)
			}
		}
		return true
	})
	if len(clip) == 0 {
		t.Fatal("scene produced no triangles")
	}
	gpuCov := gpuRasterCoverage(t, dev, clip, w, h)
	gpuN := 0
	for _, ok := range gpuCov {
		if ok {
			gpuN++
		}
	}

	// Compare silhouettes. The GPU window framebuffer's Y origin convention may be
	// flipped relative to the CPU image, which is a coordinate convention, not a
	// raster disagreement; pick the orientation that agrees better and report it,
	// so the measured delta reflects edge/coverage differences only.
	diff := coverageDiff(cpuCov, gpuCov, w, h, false)
	diffFlip := coverageDiff(cpuCov, gpuCov, w, h, true)
	flipped := diffFlip < diff
	if flipped {
		diff = diffFlip
	}
	frac := float64(diff) / float64(w*h)
	t.Logf("forward-raster coverage: cpu=%d gpu=%d differ=%d (%.2f%% of %d px, yflip=%v)",
		cpuN, gpuN, diff, frac*100, w*h, flipped)

	if cpuN == 0 || gpuN == 0 {
		t.Fatalf("one side rendered nothing (cpu=%d gpu=%d)", cpuN, gpuN)
	}
	// Step-1 loose gate: the silhouettes must broadly agree. The real (tight) gate
	// is set from the logged number once CI reports it.
	if frac > 0.10 {
		t.Fatalf("CPU/GPU forward-raster coverage differs on %.2f%% of pixels (step-1 bound 10%%)", frac*100)
	}
}

// coverageDiff counts pixels whose coverage differs between a (CPU, top-down) and
// b (GPU); when flipY, b is read bottom-up to account for the framebuffer origin.
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

const fwdCovVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _v { float pos[]; };
void main() {
	int i = gl_VertexID;
	gl_Position = vec4(pos[i*4], pos[i*4+1], pos[i*4+2], pos[i*4+3]);
}`

const fwdCovFrag = `#version 310 es
precision highp float;
out vec4 fragColor;
void main() { fragColor = vec4(1.0, 1.0, 1.0, 1.0); }`

// gpuRasterCoverage rasterizes the given clip-space triangles (4 floats/vertex,
// 3 vertices/triangle) on the GL backend with depth testing and returns a
// per-pixel coverage mask (true where a fragment was written).
func gpuRasterCoverage(t *testing.T, dev *gpu.Device, clip []float32, w, h int) []bool {
	t.Helper()
	vmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: fwdCovVert})
	if err != nil {
		t.Fatalf("vertex module: %v", err)
	}
	fmod, err := dev.NewShaderModule(gpu.ShaderSource{GLSL: fwdCovFrag})
	if err != nil {
		t.Fatalf("fragment module: %v", err)
	}
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: vmod, VertexEntry: "main",
		FragmentModule: fmod, FragmentEntry: "main",
		ColorFormat: gpu.RGBA8Unorm,
		DepthFormat: gpu.Depth32Float,
	})
	if err != nil {
		t.Fatalf("render pipeline: %v", err)
	}
	color, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA8Unorm, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		t.Fatalf("color texture: %v", err)
	}
	d, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.Depth32Float, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		t.Fatalf("depth texture: %v", err)
	}
	vbuf, err := dev.NewBuffer(gpu.BufferDescriptor{Data: floatsToBytes(clip), Usage: gpu.BufferStorage})
	if err != nil {
		t.Fatalf("vertex buffer: %v", err)
	}

	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: color, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 1},
		DepthTexture: d, ClearDepth: 1,
	})
	rp.SetPipeline(pipe)
	rp.SetVertexBuffer(0, vbuf)
	rp.Draw(gpu.TriangleList, 0, len(clip)/4)
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	pix := color.ReadPixels()
	cov := make([]bool, w*h)
	for i := 0; i < w*h; i++ {
		cov[i] = pix[i*4] > 127 // white where covered
	}
	return cov
}

func floatsToBytes(d []float32) []byte {
	b := make([]byte, len(d)*4)
	for i, f := range d {
		u := stdmath.Float32bits(f)
		b[i*4] = byte(u)
		b[i*4+1] = byte(u >> 8)
		b[i*4+2] = byte(u >> 16)
		b[i*4+3] = byte(u >> 24)
	}
	return b
}
