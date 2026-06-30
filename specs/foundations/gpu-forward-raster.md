---
title: "GPU forward rasterizer (brick 3b): scene wiring + parity-by-measurement"
status: in progress (primitives + G-buffer raster done; attribute conventions next)
depends_on:
  - foundations/gpu-render-depth.md
  - foundations/gpu-render-mrt.md
affects:
  - render
  - gpu
created: 2026-06-30
author: changkun
dispatched_task_id: null
---

# GPU forward rasterizer (brick 3b)

## Why

The forward rasterizer (vertex transform + triangle raster -> the deferred
G-buffer) is the last CPU-only stage of the renderer; the deferred shading pass
already runs on the GPU. Moving the rasterizer onto the GPU makes the renderer
fully GPU and (with the on-screen Surface seam) enables a GPU render -> present
pipeline with no CPU round-trip. Bricks 1-2 (depth, MRT) are done on Metal; brick
3a ported depth+MRT to the GL backend so this is CI-verifiable on Mesa.

## Current CPU path (what GPU must reproduce)

`render/raster.go`: `passForward` iterates scene objects; `draw` computes
`trans = Proj*View*Model`, transforms vertices to clip space, perspective
`recipw = -1/Pos.W`, viewport transform -> screen, back-face cull, clip to the
viewport AABB (`clipTriangle`/`drawClipped`), then scanline-fills a
`buffer.FragmentBuffer`: per pixel Depth (in [0,1]), Nor, Col, UV, MaterialID.
`passDeferred` then shades that G-buffer (already on GPU via
`gpuDeferredShade(dev, buf *buffer.FragmentBuffer, ...)`).

## The G-buffer seam (named, per review)

The GPU deferred pass takes a CPU `*buffer.FragmentBuffer` and uploads it to GPU
storage buffers internally. So today the G-buffer lives on the CPU. Two options
for 3b, decided when step 2 lands:
- **(A) Produce the CPU `FragmentBuffer` from the GPU raster** (read the MRT
  attachments back, fill the buffer) — minimal blast radius, keeps the deferred
  seam unchanged, but adds a GPU->CPU readback between forward and deferred.
- **(B) Keep the G-buffer on the GPU** (MRT textures) and add a deferred entry
  that reads GPU textures instead of the CPU buffer — removes the round-trip, the
  real end goal, but a bigger seam change.
Start with (A) for parity bring-up; move to (B) once parity holds.

## Parity is by measurement, not pixel-exactness (key decision)

Deferred-shading parity works because CPU and GPU shade the SAME input G-buffer.
The rasterizer is different in kind: it decides which fragments exist and their
interpolated values. A CPU scanline rasterizer and GPU hardware raster WILL
disagree at silhouette edges, depth ties, and clip boundaries (top-left fill
rules, sub-pixel/depth precision, perspective-correct interpolation rounding,
hardware vs explicit `drawClipped` clipping). Pixel-exact parity is NOT
achievable and NOT the oracle.

So the gate is tolerance-based, and the tolerance is SET BY MEASUREMENT, not
guessed: render a known scene through CPU `passForward` and the GPU forward
raster, dump the per-pixel difference distribution (interior vs edge band, max
and percentile deltas), and pick the gate from what is observed. Template:
`render/gl_render_linux_test.go` (`TestGLDeferredRender`: CPU vs GPU(GL), asserts
`<2% of channels differ by >8`). Likely shape: interior pixels within epsilon +
a bounded fraction allowed to differ (the edge band), plus a golden-stability
test on the GPU output itself.

## Characterization of the GPU vs CPU G-buffer (measured on Mesa, brick 3b)

All GPU primitives exist and are CI-exact: coverage raster, GPU vertex transform
(coverage pixel-identical), depth, MRT, RGBA32Float targets. The float G-buffer
raster runs. Measuring its attributes against the CPU forward pass
(`buf.UnsafeGet`) shows three divergences, all rooted in CPU quirks, NOT GPU bugs
(so EXACT parity is impossible; tolerance-by-measurement is required, as planned):
- **Normal** (mean ~0.88 after back-face culling): the per-vertex world normals
  are identical (verified: CPU-computed vs in-shader matrix give bit-identical
  results). The divergence is INTERPOLATION -- GLSL varyings are perspective-
  correct, but CPU `drawClipped` interpolates normals LINEARLY (plain barycentric,
  no recipw). With this close/wide-FoV scene that is large. The GPU normal is the
  more-correct one. GLSL ES has no `noperspective` qualifier to match the CPU.
- **World position** (mean ~0.53): the CPU `drawClipped` has a BUG --
  `pos = (v0.worldX, v1.worldY, v2.worldZ)`, a per-triangle constant (lines
  ~510-515). The GPU computes correct interpolated world pos.
- **Depth** (mean ~0.95): pure encoding offset, ordering identical -- CPU stores
  ndc_z in [-1,1] (~-0.9 near), GPU `gl_FragCoord.z` is (ndc_z+1)/2 in [0,1].
  Remap with `2*z-1` when populating the FragmentBuffer.
- Back-face culling matched via `gl_FrontFacing` discard (the position negation
  preserves NDC winding, GL CCW-front matches the CPU screen cross-z>0).

DECISION NEEDED (product call): for a drop-in GPU forward pass, either (A)
replicate the CPU quirks bug-for-bug (linear normal interp via a w-trick;
per-triangle worldpos) for exact image parity, or (B) ship the more-correct GPU
G-buffer and gate the FINAL deferred-shaded image with a measured tolerance
(accepting it differs from -- and improves on -- the CPU). (B) is cleaner but
changes output; (A) preserves current output incl. a known bug.

## Steps

1. **Measurement — DONE.** `render/gpu_forward_measure_linux_test.go`
   (`TestGPUForwardRasterCoverage`): rasterize the scene's triangles in identical
   SCREEN space (computed CPU-side exactly as `draw()`, mapped to NDC w=1 to dodge
   the projection's negated-W) on the GL backend and compare silhouette coverage
   to the CPU `passForward`. RESULT (Mesa llvmpipe, in CI): cpu=3220 gpu=3220
   differ=0 (0.00%, no Y-flip). So COVERAGE is pixel-exact given identical
   screen-space input on the CI oracle; the coverage gate can be tight. The
   parity deltas to worry about are in ATTRIBUTE INTERPOLATION + depth precision,
   measured in step 2 -- not coverage.
2. **Full G-buffer raster.** Vertex shader (`trans*pos`, normal/worldpos/uv
   varyings) + fragment writing the MRT G-buffer (normal, worldpos, basecol,
   matid) + depth; produce the `FragmentBuffer` (seam option A). Gate by
   `passGPU["forward"]`, parity vs CPU at the measured tolerance, feeding the
   existing GPU deferred shading. Backend-agnostic: GL is the CI oracle, Metal is
   the darwin runtime.
3. **Remove the round-trip (seam option B)** once parity holds: keep the G-buffer
   on GPU textures into the deferred pass.

## Out of scope

- MSAA on the GPU raster (the CPU path supersamples; match at MSAA=1 first).
- Texture-sampled materials in the GPU G-buffer (basecol from texture) until flat
  materials parity holds.
