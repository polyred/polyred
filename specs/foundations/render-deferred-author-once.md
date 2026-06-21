---
title: "Render deferred: use the author-once kernel on the GPU"
status: drafted
depends_on:
  - foundations/author-once-kernels.md
  - foundations/render-pass-runner.md
  - foundations/gpu-by-default.md
affects:
  - render/gpudeferred.go
  - gpu/shader/gpumath/kernels/shade.go
  - render/gpudeferred_multimat_test.go
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Render deferred: use the author-once kernel on the GPU

## Overview

The fourth bounded slice of the unified renderer ([unified-renderer.md](
unified-renderer.md)). Today the renderer's GPU deferred pass compiles a private
`deferredKernel` string (in `render/gpudeferred.go`) that duplicates the
Blinn-Phong logic also held, author-once, in `gpu/shader/gpumath/kernels`
(`Shade`/`ShadeSrc`). This slice deletes the duplicate: the renderer's GPU
deferred pass compiles `kernels.ShadeSrc`, so the GPU shader and the standalone
parity harness come from one source.

Scope is the GPU path only. The CPU default still runs
`shader.FragmentShader`; folding the CPU default onto `kernels.Shade` is a
separate, riskier slice (it can shift golden output) and is out of scope here.
Accurate framing: after this slice, "GPU deferred uses the author-once kernel,
and CPU-run-as-Go of that same kernel is proven to match it"; not yet "CPU and
GPU shading from one source".

## Current State

- `render/gpudeferred.go` holds `const deferredKernel` (a Go-DSL string with a
  `Scene` struct uniform) and `runDeferredKernel`, which compiles it to MSL,
  binds binding 6 (`scene`) as a `UniformBuffer`, and dispatches.
- The scene buffer is already a `[]float32`:
  `{camPos.X, camPos.Y, camPos.Z, 1, ambientI, numLights, 0, 0}` (gpudeferred.go
  ~line 362), the exact layout `kernels.Shade` reads.
- `gpu/shader/gpumath/kernels/shade.go` holds the author-once `Shade`; `embed.go`
  exposes its source as `ShadeSrc`. It is proven GPU == CPU-as-Go == reference in
  `gpu/parity_shared_test.go` (scene buffer there uses `scene[3]=0`).
- `deferredSelfCheck` (gpudeferred.go) re-derives the shading with a hand-written
  pure-Go replica and `println`s a mismatch; it is gated by the package var
  `debugDeferredSelfCheck`, toggled on in `gpudeferred_multimat_test.go`.
- `assertDeferredClose` (gpudeferred_compare_test.go) fails when >2% of channels
  differ by >8 between the CPU render and the GPU render. This is the real guard
  on GPU-shading correctness against `shader.FragmentShader`.

## Components

### `kernels.Shade` view-vector W fix

`shade.go` builds the camera position as `V4(scene[0], scene[1], scene[2], 0)`.
The hardcoded `0` makes `camPos.Sub(wpos)` carry `W = 0 - wpos.W`, which pollutes
`Normalize` (the view vector) whenever the caller's `wpos.W != 0`. The standalone
parity scene uses `scene[3]=0` and `wpos.W=0`, so it never exposes this; the
renderer uses `scene[3]=1` and `wpos.W=1`, where `W=0` corrupts the view vector
by ~10-30% at typical camera distances.

Fix: read W from the buffer, `V4(scene[0], scene[1], scene[2], scene[3])`. This
adapts the one kernel to both conventions: parity keeps `W=0` (unchanged), the
renderer gets `W=1` (matching `deferredKernel`'s `s.CamPos.W = scene[3]`).

### `runDeferredKernel` swap

- Compile `kernels.ShadeSrc` instead of the local `deferredKernel`.
- Bind binding 6 (`scene`) as a `StorageBuffer` (`kernels.Shade` takes
  `scene []float32`, not a struct uniform): change the layout entry to `sb(6)`
  and `scb` from `uniformBuf` to `storageBuf`. The float values are identical, so
  GPU output is numerically unchanged versus the old uniform path.
- Add the import `poly.red/gpu/shader/gpumath/kernels`.
- Delete `const deferredKernel` (its only references are this comment and
  `runDeferredKernel`; no byte-golden MSL test pins it).

### `deferredSelfCheck` uses `kernels.Shade`

Replace the hand-written replica with calls to `kernels.Shade` (run as Go) over
the same G-buffer, for each `okMask && !passthrough` fragment, into a replica
output buffer; compare to the GPU output. This makes the gated debug check prove
"GPU(`ShadeSrc`) == `kernels.Shade`-as-Go" directly (compiler-lowering check),
rather than against a separate replica that can silently drift.

## Testing Strategy

The W fix is a real bug fix and needs fails-without / passes-with evidence. The
self-check does NOT provide it: both its sides derive from the same `shade.go`,
so it passes for any W. The guard that flips is `assertDeferredClose` (CPU
`shader.FragmentShader` vs GPU `kernels.Shade`) in the deferred and multi-material
tests.

- Empirically verify the flip on darwin (Metal), `GOWORK=off`:
  1. After the `runDeferredKernel` swap but with `kernels.Shade` still at `W=0`,
     run the deferred/multi-material tests; confirm they FAIL (the GPU view vector
     is wrong vs the CPU reference).
  2. Apply the `scene[3]` fix; confirm they PASS.
  Record both outcomes; that run is the regression evidence.
- Promote the self-check from `println` to a hard assertion: have
  `deferredSelfCheck` record its match result in a package var and have
  `TestGPUDeferredMultiMaterial` fail on mismatch. This locks "GPU deferred ==
  author-once kernel" so a future kernel edit that breaks lowering is caught.
- Standalone parity (`gpu/parity_shared_test.go`) stays green unchanged
  (`scene[3]=0` keeps `W=0`).
- Existing golden/deferred image tests stay green (output numerically unchanged
  by the uniform to storage rebind).

## Out of scope (separate bounded specs)

- Folding the CPU default (`shader.FragmentShader`) onto `kernels.Shade` (4b):
  riskier, may shift golden output; needs its own reconciliation slice.
- Multi-backend render (compile `ShadeSrc` to GLSL/SPIR-V in render, lift the
  Metal-only restriction).
- Shadow and AO kernels onto the author-once path.

## Deliverable

`render/gpudeferred.go` GPU deferred compiles `kernels.ShadeSrc`; the duplicate
`deferredKernel` is gone; `kernels.Shade`'s view-vector W reads from the buffer;
the multi-material test hard-asserts GPU == author-once kernel; CI green on macOS
(Metal) and Linux/Windows (all-CPU, unchanged).
