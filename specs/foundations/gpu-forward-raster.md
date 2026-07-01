---
title: "GPU forward rasterizer (brick 3b): scene wiring + parity-by-measurement"
status: DONE -- GPU forward is the default passForward on GL (CI) and Metal (darwin), gated by measured parity
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
- **World position** (mean ~0.53): the CPU `drawClipped` worldpos BUG
  (`pos = (m1.X, m2.Y, m3.Z)`) is now FIXED (interpWorldPos + TestInterpWorldPos;
  goldens still pass). The residual delta is the SAME quirk as normals -- the CPU
  interpolates worldpos linearly, the GPU perspective-correct.
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
2. **Full G-buffer raster + deferred integration — DONE (CI-proven on GL).**
   `TestGPUForwardDeferredIntegration`: the GPU forward raster's G-buffer
   (world normal + position, RGBA32F MRT, depth-tested, back-face culled via
   gl_FrontFacing) is injected into the renderer's FragmentBuffer and the real
   deferred shading pass runs on the GL device; the final image matches the
   all-CPU render at **0.00% of channels differing by >16** (gated at 2%). The
   perspective-correct (GPU) vs linear (CPU) normal/worldpos interpolation only
   produces sub-quantization lighting differences, so the shaded image is
   equivalent. The CPU worldpos bug was fixed (interpWorldPos) so CPU+GPU agree.
   REMAINING to fully finish brick 3b:
   - Wire it as `passGPU["forward"]` IN the renderer (a real `gpuForwardRaster`
     emitting the G-buffer incl. matid/col, gated like the deferred pass), so the
     renderer uses the GPU forward by default -- today it is proven via a
     white-box test that injects the G-buffer, not yet the default path.
   - Port the forward raster to the Metal backend (darwin runtime; GL is the CI
     oracle), as the deferred pass already is.
3. **Remove the round-trip (seam option B)** once wired: keep the G-buffer on GPU
   textures into the deferred pass (no CPU FragmentBuffer round-trip).

## Wired as the default (2026-07-01): the Y-flip bug and the parity band

`passForward` now dispatches `runPass("forward", gpuForwardPass, cpuForwardPass)`:
the GL device rasterizes the full G-buffer (world position, normal, uv + du/dv
mipmap-LOD gradients via dFdx/dFdy, material id, depth) by default; the CPU is the
fallback when there is no device or it cannot run the GLSL pipeline (Metal errors on
missing MSL, so darwin/goldens are unchanged).

Bringing it up surfaced ONE real data-path bug and then the expected parity band:

- **Y-flip (fixed).** Render-target *texture* readback follows GL's bottom-left
  origin, so `glReadPixels` row r is screen row h-1-r. `gpuForwardPass` wrote
  readback row y to FragmentBuffer row y, vertically mirroring the whole G-buffer --
  every fragment sampled the texture at the mirror pixel (meanU 0.21 / meanV 0.34;
  final image 24.9% @>8). The deferred pass reads a *compute SSBO* (linear, not
  flipped), which is why `TestGLDeferredRender` never caught it. Found by overlap
  measurement: overlap(cpu, gpuForward) flip=1811/1811 vs noflip=950. Fix: read the
  mirrored source row when filling the buffer. (Metal/other texture-readback callers
  should assume the same bottom-origin convention.)

- **Parity band (accepted, gated by measurement).** After the fix the residual vs
  the all-CPU render is deterministically **4.38% @>8 / 0.97% @>16** (96x96 bunny).
  Attribution: substituting the CPU's Nor/WordPos into the GPU G-buffer leaves >8
  *unchanged*, so the residual is 100% UV, not the perspective-vs-linear normal/
  worldpos (that washes out, as measured). Interior split: of ~1600 interior
  fragments the smooth surface is UV-clean (same-triangle meanU/V ~0), and ~4% are
  large-magnitude UV diffs concentrated at internal folds -- where GPU (hardware
  hyperbolic depth) and CPU (float barycentric depth) pick different but coincident
  triangles at a depth tie. That plus the silhouette edge band is the parity trap
  this spec predicted; the GPU is not wrong, it makes an equally-valid choice at
  ties. Per the user's "ship correct GPU + tolerance" call, the gate is measured,
  not zero.

## Metal port (2026-07-01): runs on the darwin runtime too

`gpuForwardPass` carries MSL shaders (`fwdGBufMSL`) beside the GLSL, so `passForward`
runs fully on the GPU on darwin (Metal), not only GL in CI. Both shader modules carry
the MSL library; GL ignores the pipeline entry names (GLSL is always `main`), Metal
selects `fwdVert`/`fwdFrag`. Three Metal conventions differ from GL and were each found
empirically against the CPU (all fixed in the MSL; the GLSL is unchanged):

- **Clip z range.** The renderer's projection yields GL-style ndc z in [-1,1], but
  Metal clips z to [0,1] and dropped the near half (coverage 0). The MSL vertex remaps
  `z' = (z + w)/2`; the fragment still recovers the CPU's [-1,1] via `position.z*2-1`.
  The deferred pass never hit this because it is a compute kernel (no clip).
- **Back-face winding is inverted.** Metal's default front-facing winding is the
  opposite of GL's for this NDC geometry, so the MSL discards `front` where the GLSL
  discards `!gl_FrontFacing` (both keep the CPU's front faces). The wrong sense keeps
  back faces: correct silhouette, wrong per-pixel data, ~15% off. No Y-flip is needed
  on Metal despite its top-left texture origin (the readback here is bottom-origin like
  GL, confirmed by coverage overlap = CPU exactly).
- **RGBA32Float texture readback.** `metalTexture.readPixels` hardcoded 4 bytes/pixel;
  a float G-buffer target needs 16. Added a bytes-per-pixel field set from the format,
  mirroring the GL backend's float-aware readback.

Result on Metal: coverage == CPU exactly, final image 4.37%@>8 / 0.97%@>16 -- the same
parity band as GL. `TestGPUForwardMetal` (render/gpu_forward_darwin_test.go) gates it.

Wiring GPU-forward as the default silently widened every full-`Render()` GPU-vs-CPU
gate to include the forward parity band; the darwin deferred/gamma parity tests (tight
exactness bound) then failed. Fixed with an unexported `forwardCPU` config flag +
test-only `forwardOnCPU()` option: those single-pass gates force the CPU forward and
shade the same G-buffer as the CPU reference (same isolation as `TestGLDeferredRender`).

Three gates lock the GL path in (all GL, Mesa surfaceless, in the gl-probe run filter):
- `TestGPUForwardPassUV` -- confound-free forward gate (no shading/AA): interior
  same-triangle UV must agree to ~float precision (<0.01, catches any interpolation
  regression at the source, e.g. a re-introduced Y-flip); interior different-triangle
  (fold/seam) fraction bounded <8% (measured ~4.2%).
- `TestGPUForwardDeferredIntegration` -- full pipeline (GPU forward -> deferred -> AA
  vs all-CPU), gated <6% @>8 (measured 4.38%; @>8 not @>16, so the 8-16 band a subtle
  regression would first show in is not blind).
- `TestGLDeferredRender` -- now CPU forward + GPU deferred, keeping the pure
  deferred-shading gate tight (<2% @>8) independent of the forward parity band.

## Out of scope

- MSAA on the GPU raster (the CPU path supersamples; match at MSAA=1 first).
- Texture-sampled materials in the GPU G-buffer (basecol from texture) until flat
  materials parity holds.
