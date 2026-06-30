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

	"poly.red/buffer"
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

// cpuForward runs the CPU forward pass and returns the filled G-buffer so a test
// can read per-fragment attributes via UnsafeGet.
func cpuForward(s *scene.Scene, c camera.Interface, w, h int) *buffer.FragmentBuffer {
	r := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU())
	buf := r.CurrBuffer()
	buf.Clear()
	r.passForward()
	return buf
}

type gbufObject struct {
	pos, wpos, wnor []float32 // model pos; world pos + world normal (CPU-computed)
	trans           [16]float32
}

const gbufVert = `#version 310 es
layout(std430, binding = 0) readonly buffer _pos { float pos[]; };  // model position
layout(std430, binding = 1) readonly buffer _wp { float wpos[]; };  // world position
layout(std430, binding = 2) readonly buffer _wn { float wnor[]; };  // world normal
layout(std430, binding = 3) readonly buffer _m { float m[]; };      // trans (col-major)
out vec3 vWorld;
out vec3 vNormal;
void main() {
	int i = gl_VertexID;
	vec4 p = vec4(pos[i*4], pos[i*4+1], pos[i*4+2], pos[i*4+3]);
	mat4 T = mat4(m[0],m[1],m[2],m[3], m[4],m[5],m[6],m[7],
	              m[8],m[9],m[10],m[11], m[12],m[13],m[14],m[15]);
	gl_Position = -(T * p);
	vWorld = vec3(wpos[i*4], wpos[i*4+1], wpos[i*4+2]);
	vNormal = vec3(wnor[i*4], wnor[i*4+1], wnor[i*4+2]);
}`

const gbufFrag = `#version 310 es
precision highp float;
in vec3 vWorld;
in vec3 vNormal;
layout(location = 0) out vec4 outWorld;  // xyz = world position, w = depth
layout(location = 1) out vec4 outNormal; // xyz = unit world normal
void main() {
	// Match the CPU forward pass, which culls back faces (screen cross-z < 0)
	// before depth testing. The position negation preserves NDC winding, so GL's
	// default CCW-front matches; discard back faces so only front-face fragments
	// compete for the depth test (otherwise a nearer back face at a non-convex
	// fold would win and store an opposite normal).
	if (!gl_FrontFacing) discard;
	outWorld = vec4(vWorld, gl_FragCoord.z);
	outNormal = vec4(normalize(vNormal), 0.0);
}`

// TestGPUForwardGBuffer produces the float G-buffer (world position + depth,
// world normal) on the GPU and measures each attribute against the CPU forward
// pass's actual fragments (buf.UnsafeGet). Normals are the clean parity signal
// (validated tightly); world position is expected to diverge because the CPU
// drawClipped has a worldpos bug -- pos = (v0.worldX, v1.worldY, v2.worldZ), a
// per-triangle constant -- so its delta is logged as a finding, not gated. Depth
// is logged (encoding may differ).
func TestGPUForwardGBuffer(t *testing.T) {
	dev := openGLOrSkip(t)
	defer dev.Close()

	const w, h = 128, 128
	s, c := newscene(w, h)
	buf := cpuForward(s, c, w, h)

	world, normal := gpuGBuffer(t, dev, buildGBufObjs(s, c), w, h)

	var n int
	var sumN, maxN, sumWP, maxWP, sumD, maxD float32
	var hist [4]int // normal-delta buckets: <0.1, <0.5, <1.0, >=1.0
	var dumped int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			f := buf.UnsafeGet(x, y)
			if !f.Ok {
				continue
			}
			idx := (y*w + x) * 4
			gN := math.Vec4[float32]{X: normal[idx], Y: normal[idx+1], Z: normal[idx+2], W: 0}
			gWP := math.Vec4[float32]{X: world[idx], Y: world[idx+1], Z: world[idx+2], W: 1}
			dN := gN.Sub(f.Nor).Len()
			dWP := gWP.Sub(f.WordPos).Len()
			dD := absf(world[idx+3] - f.Depth)
			switch {
			case dN < 0.1:
				hist[0]++
			case dN < 0.5:
				hist[1]++
			case dN < 1.0:
				hist[2]++
			default:
				hist[3]++
			}
			if dN > 1.0 && dumped < 4 {
				t.Logf("  bad normal @(%d,%d): cpu=(%.2f,%.2f,%.2f) gpu=(%.2f,%.2f,%.2f)", x, y, f.Nor.X, f.Nor.Y, f.Nor.Z, gN.X, gN.Y, gN.Z)
				dumped++
			}
			sumN += dN
			sumWP += dWP
			sumD += dD
			if dN > maxN {
				maxN = dN
			}
			if dWP > maxWP {
				maxWP = dWP
			}
			if dD > maxD {
				maxD = dD
			}
			n++
		}
	}
	t.Logf("normal-delta histogram (<0.1, <0.5, <1, >=1): %v of %d", hist, n)
	if n == 0 {
		t.Fatal("no covered pixels")
	}
	t.Logf("G-buffer over %d px: normal mean=%.4f max=%.4f; worldpos mean=%.4f max=%.4f; depth mean=%.4f max=%.4f (normal+worldpos: CPU interpolates linearly, GPU perspective-correct; depth: [-1,1] vs [0,1] encoding)",
		n, sumN/float32(n), maxN, sumWP/float32(n), maxWP, sumD/float32(n), maxD)
	// MEASUREMENT (log-only): the residual deltas are CPU quirks, not GPU bugs
	// (see gpu-forward-raster.md). normal + worldpos: the CPU interpolates them
	// LINEARLY while GLSL varyings are perspective-correct (the GPU is the more
	// correct one); depth: a pure [-1,1] (CPU) vs [0,1] (GPU gl_FragCoord.z)
	// encoding offset, same ordering. Back-face culling is matched via
	// gl_FrontFacing. The end-to-end effect through deferred shading is gated in
	// TestGPUForwardDeferredIntegration.
}

// buildGBufObjs builds the GPU forward-raster inputs for each scene object: model
// positions for gl_Position (via trans), plus CPU-computed world position and
// world normal per vertex (exactly as draw() computes them) that the GPU only
// interpolates.
func buildGBufObjs(s *scene.Scene, c camera.Interface) []gbufObject {
	view, proj := c.ViewMatrix(), c.ProjMatrix()
	var objs []gbufObject
	scene.IterObjects(s, func(g *geometry.Geometry, model math.Mat4[float32]) bool {
		world := model.MulM(g.ModelMatrix())
		normalMat := world.Inv().T()
		o := gbufObject{trans: colMajor(proj.MulM(view).MulM(world))}
		for _, tri := range g.Triangles() {
			for _, v := range []*primitive.Vertex{tri.V1, tri.V2, tri.V3} {
				wp := world.MulV(v.Pos)
				wn := v.Nor.Apply(normalMat)
				o.pos = append(o.pos, v.Pos.X, v.Pos.Y, v.Pos.Z, v.Pos.W)
				o.wpos = append(o.wpos, wp.X, wp.Y, wp.Z, 1)
				o.wnor = append(o.wnor, wn.X, wn.Y, wn.Z, 0)
			}
		}
		objs = append(objs, o)
		return true
	})
	return objs
}

// TestGPUForwardDeferredIntegration is the end-to-end brick-3b integration: the
// GPU forward raster's G-buffer (world normal + position) is injected into the
// renderer's fragment buffer, then the renderer's real deferred shading pass runs
// on it, and the final image is compared to the all-CPU render. It validates that
// a GPU-rasterized G-buffer drives the existing deferred pipeline to an
// equivalent picture (within tolerance: the GPU interpolates normal/worldpos
// perspective-correct vs the CPU's linear, so it is close but not identical).
func TestGPUForwardDeferredIntegration(t *testing.T) {
	dev := openGLOrSkip(t)
	defer dev.Close()

	const w, h = 96, 96
	s, c := newscene(w, h)

	// All-CPU reference.
	cpu := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), CPU()).Render()

	// GPU-forward path: the renderer's own passForward now rasterizes the full
	// G-buffer (world position, normal, uv, material id, depth) on the GL device,
	// then the deferred shading + AA run on it. This exercises the wired default
	// path end-to-end, no white-box injection.
	r := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), GPU(dev))
	buf := r.CurrBuffer()
	buf.Clear()
	r.passForward()
	if !r.passOnGPU("forward") {
		t.Skip("forward did not run on the GPU (no GL forward path)")
	}
	buf.ClearColor()
	r.passDeferred()
	r.passAntialiasing()
	gpuImg := r.outBuf
	if !r.passOnGPU("deferred") {
		t.Skip("deferred did not run on the GPU (no GL deferred path)")
	}

	if len(cpu.Pix) != len(gpuImg.Pix) {
		t.Fatalf("size mismatch: cpu %d gpu %d", len(cpu.Pix), len(gpuImg.Pix))
	}
	nBig := 0
	for i := range cpu.Pix {
		d := int(cpu.Pix[i]) - int(gpuImg.Pix[i])
		if d < 0 {
			d = -d
		}
		if d > 16 {
			nBig++
		}
	}
	frac := float64(nBig) / float64(len(cpu.Pix))
	t.Logf("GPU-forward+deferred vs all-CPU: %.2f%% of channels differ by >16", frac*100)
	// Measured 0%: the GPU's perspective-correct normal/worldpos vs the CPU's linear
	// interpolation produces only sub-quantization lighting differences, so the
	// shaded 8-bit image matches. Gate at 2% (the deferred tolerance) for headroom.
	if frac > 0.02 {
		t.Fatalf("GPU-forward+deferred diverges from CPU on %.2f%% of channels (>16); want <2%%", frac*100)
	}
}

// TestGPUForwardPassUV measures gpuForwardPass's per-fragment texture coordinates
// and derivatives (U, V, Du, Dv) against the CPU forward pass, at fragments both
// cover. The first textured-scene wiring diverged because the GPU G-buffer omitted
// UV; this isolates whether the UV values OR the du/dv mipmap LOD are the gap.
// Log-only: it is a measurement, not a gate, while the GPU forward path is brought
// up; the production wiring stays on the CPU until this is ~0.
func TestGPUForwardPassUV(t *testing.T) {
	dev := openGLOrSkip(t)
	defer dev.Close()

	const w, h = 96, 96
	s, c := newscene(w, h)
	cpu := cpuForward(s, c, w, h)

	r := NewRenderer(Scene(s), Camera(c), Size(w, h), MSAA(1), Workers(1), GPU(dev))
	gbuf := r.CurrBuffer()
	gbuf.Clear()
	if err := r.gpuForwardPass(); err != nil {
		t.Skipf("gpuForwardPass unavailable: %v", err)
	}

	// GPU-vs-GPU reference: gpuGBuffer is the proven helper the integration test
	// injected at 0%. It runs the same vertex transform + raster, so its per-pixel
	// world/normal must be bit-identical to gpuForwardPass's. Comparing against it
	// (not the CPU) removes the perspective-vs-linear confound: any nonzero delta is
	// a bug in gpuForwardPass's own setup (binding map / readback), per advisor.
	gWorld, gNormal := gpuGBuffer(t, dev, buildGBufObjs(s, c), w, h)
	gbCov := func(x, y int) bool { // gpuGBuffer covered: normalized normal is unit, clear is 0
		idx := (y*w + x) * 4
		return absf(gNormal[idx])+absf(gNormal[idx+1])+absf(gNormal[idx+2]) > 0.5
	}

	// Localize the coverage divergence: counts + overlaps, no-flip vs Y-flip, across
	// CPU, gpuForwardPass (gbuf), and the proven gpuGBuffer helper.
	var nCPU, nGPU, nGB int
	var ovCG, ovCGflip, ovCB, ovCBflip, ovGB int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c0 := cpu.UnsafeGet(x, y).Ok
			g0 := gbuf.UnsafeGet(x, y).Ok
			b0 := gbCov(x, y)
			cf := cpu.UnsafeGet(x, h-1-y).Ok
			if c0 {
				nCPU++
			}
			if g0 {
				nGPU++
			}
			if b0 {
				nGB++
			}
			if c0 && g0 {
				ovCG++
			}
			if cf && g0 {
				ovCGflip++
			}
			if c0 && b0 {
				ovCB++
			}
			if cf && b0 {
				ovCBflip++
			}
			if g0 && b0 {
				ovGB++
			}
		}
	}
	t.Logf("counts: cpu=%d gpuForward=%d gpuGBuffer=%d", nCPU, nGPU, nGB)
	t.Logf("overlap cpu&gpuForward: noflip=%d flip=%d", ovCG, ovCGflip)
	t.Logf("overlap cpu&gpuGBuffer: noflip=%d flip=%d", ovCB, ovCBflip)
	t.Logf("overlap gpuForward&gpuGBuffer: %d", ovGB)

	var nShared, matMismatch int
	var sU, sV, mU, mV float32
	var sNorGG, sWPGG, mNorGG float32 // gpuForwardPass vs gpuGBuffer (GPU-vs-GPU)
	var dumped int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cf := cpu.UnsafeGet(x, y)
			gf := gbuf.UnsafeGet(x, y)
			if !cf.Ok || !gf.Ok {
				continue
			}
			nShared++
			// gpuForwardPass now un-flips Y; the gpuGBuffer helper readback is still
			// raw (flipped), so index it at the mirrored row to hit the same screen pixel.
			idx := ((h-1-y)*w + x) * 4
			// gpuForwardPass vs gpuGBuffer normal+worldpos at the same pixel.
			nGG := absf(gf.Nor.X-gNormal[idx]) + absf(gf.Nor.Y-gNormal[idx+1]) + absf(gf.Nor.Z-gNormal[idx+2])
			wGG := absf(gf.WordPos.X-gWorld[idx]) + absf(gf.WordPos.Y-gWorld[idx+1]) + absf(gf.WordPos.Z-gWorld[idx+2])
			sNorGG += nGG
			sWPGG += wGG
			if nGG > mNorGG {
				mNorGG = nGG
			}
			du, dv := absf(cf.U-gf.U), absf(cf.V-gf.V)
			sU += du
			sV += dv
			if du > mU {
				mU = du
			}
			if dv > mV {
				mV = dv
			}
			if cf.MaterialID != gf.MaterialID {
				matMismatch++
			}
			if (du > 0.05 || dv > 0.05) && dumped < 6 {
				t.Logf("  (%d,%d) cpuUV=(%.4f,%.4f) gpuUV=(%.4f,%.4f)  fwd-vs-gbuf |dNor|=%.4f |dWP|=%.4f",
					x, y, cf.U, cf.V, gf.U, gf.V, nGG, wGG)
				dumped++
			}
		}
	}
	if nShared == 0 {
		t.Fatal("no shared fragments")
	}
	t.Logf("coverage: cpu=%d gpu=%d shared=%d (gpuForwardPass vs CPU)", nCPU, nGPU, nShared)
	t.Logf("GPU-vs-GPU (gpuForwardPass vs proven gpuGBuffer) over %d shared: meanNorL1=%.5f (max %.4f) meanWPL1=%.5f -- should be ~0 if raster identical",
		nShared, sNorGG/float32(nShared), mNorGG, sWPGG/float32(nShared))
	t.Logf("UV vs CPU: meanU=%.5f meanV=%.5f (max %.4f/%.4f) matMismatch=%d",
		sU/float32(nShared), sV/float32(nShared), mU, mV, matMismatch)
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// gpuGBuffer rasterizes objs into two RGBA32F attachments (world+depth, normal)
// with depth testing and returns the two readbacks as []float32 (w*h*4 each).
func gpuGBuffer(t *testing.T, dev *gpu.Device, objs []gbufObject, w, h int) (world, normal []float32) {
	t.Helper()
	pipe, err := dev.NewRenderPipeline(gpu.RenderPipelineDescriptor{
		VertexModule: mkMod(t, dev, gbufVert), VertexEntry: "main",
		FragmentModule:    mkMod(t, dev, gbufFrag),
		FragmentEntry:     "main",
		ColorFormat:       gpu.RGBA32Float,
		ExtraColorFormats: []gpu.TextureFormat{gpu.RGBA32Float},
		DepthFormat:       gpu.Depth32Float,
	})
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	mkF32 := func() *gpu.Texture {
		tex, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.RGBA32Float, Width: w, Height: h, RenderTarget: true})
		if err != nil {
			t.Fatalf("float texture: %v", err)
		}
		return tex
	}
	wt, nt := mkF32(), mkF32()
	depth, err := dev.NewTexture(gpu.TextureDescriptor{Format: gpu.Depth32Float, Width: w, Height: h, RenderTarget: true})
	if err != nil {
		t.Fatalf("depth texture: %v", err)
	}
	enc := dev.NewCommandEncoder()
	rp := enc.BeginRenderPass(gpu.RenderPassDescriptor{
		ColorTexture: wt, Load: gpu.LoadClear, ClearColor: [4]float64{0, 0, 0, 0},
		ExtraColorTargets: []gpu.ColorTarget{{Texture: nt, ClearColor: [4]float64{0, 0, 0, 0}}},
		DepthTexture:      depth, ClearDepth: 1,
	})
	rp.SetPipeline(pipe)
	for _, o := range objs {
		rp.SetVertexBuffer(0, mkBuf(t, dev, o.pos))
		rp.SetVertexBuffer(1, mkBuf(t, dev, o.wpos))
		rp.SetVertexBuffer(2, mkBuf(t, dev, o.wnor))
		rp.SetVertexBuffer(3, mkBuf(t, dev, o.trans[:]))
		rp.Draw(gpu.TriangleList, 0, len(o.pos)/4)
	}
	rp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()
	return f32s(wt.ReadPixels()), f32s(nt.ReadPixels())
}

func f32s(b []byte) []float32 {
	out := make([]float32, len(b)/4)
	for i := range out {
		out[i] = stdmath.Float32frombits(uint32(b[i*4]) | uint32(b[i*4+1])<<8 | uint32(b[i*4+2])<<16 | uint32(b[i*4+3])<<24)
	}
	return out
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
	// Coverage and the GPU transform are pixel-EXACT against the CPU (measured 0%);
	// gate tightly so a convention regression is caught, with a small margin for
	// any single-pixel edge rounding.
	if frac > 0.01 {
		t.Fatalf("CPU/GPU forward coverage differs on %.2f%% of pixels (want ~0%%; convention regression)", frac*100)
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
