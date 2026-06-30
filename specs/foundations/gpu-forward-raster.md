---
title: "GPU forward rasterizer (brick 3b): scene wiring + parity-by-measurement"
status: in progress (step 1: measurement)
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

## Steps

1. **Measurement (this step).** A minimal GPU forward raster of the scene
   geometry producing one comparable attribute (start with DEPTH and/or
   world-space normal), rendered on the GL backend (3a depth/MRT), compared
   against the CPU `passForward` output for the same attribute. Dump the delta
   distribution (CI log) and assert a deliberately loose bound first; tighten to
   the measured tolerance once the distribution is known. This calibrates the
   gate before the full G-buffer exists. CI: a new `render` test, Mesa
   surfaceless, in the gl-probe filter.
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
