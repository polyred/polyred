---
title: "Lock CPU default shading to the author-once kernel"
status: implemented (CI-pending)
depends_on:
  - foundations/render-deferred-author-once.md
affects:
  - render/shading_equiv_test.go
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Lock CPU default shading to the author-once kernel

## Overview

The fifth bounded slice of the unified renderer ([unified-renderer.md](
unified-renderer.md)), and the faithful completion of "CPU and GPU share one
shading abstraction". After [render-deferred-author-once.md](
render-deferred-author-once.md), the GPU deferred pass shades through the
author-once `kernels.Shade`. The CPU default still shades through
`shader.FragmentShader`. This slice does NOT merge them into one function; it
proves they are equivalent and pins that with a test.

## Why not merge

Merging the CPU default onto `kernels.Shade` was considered and rejected:

- `kernels.Shade` is slice-based (built for GPU buffer layout). The CPU deferred
  path is per-fragment (`DrawFragments -> shade -> FragmentShader`) and is the
  PRIMARY path on Linux/Windows (the render GPU offload is Metal-only). Calling a
  slice kernel per fragment allocates per pixel: a real regression on the primary
  path. Batching it into a G-buffer loop is a much larger refactor.
- `shader.FragmentShader` cannot be deleted regardless: it owns texture/lod,
  `FlatShading -> FaceNor`, the no-lights early return, and it is the fallback for
  every `errGPUDeferredUnsupported` scene.

So "one Blinn-Phong function" is neither achievable nor desirable. The real goal,
"share the same abstraction", is met by making `kernels.Shade` the single source
of truth and proving the CPU default agrees with it.

## Components

### `render/shading_equiv_test.go`

A pure-CPU test (no device, runs on every platform) that, for a representative
material and a spread of G-buffer fragments (varied normals/positions, point +
directional + ambient lights):

- shades via `shader.FragmentShader` (the CPU default), and
- shades via `kernels.Shade` run as Go, marshaling inputs with the exact layout
  of `render/gpudeferred.go` (lights, materials, scene = `[camPos.xyz, 1,
  ambientI, numLights, ...]`),
- quantizes the kernel output with the SAME `Round`+clamp the CPU path bakes in,
- asserts every channel agrees within 1 LSB.

The 1-LSB bound is intentional: the only intrinsic difference is float
accumulation order (`FragmentShader` factors `Diffuse` out of the light sum; the
kernel distributes it per light). A wider gap means a real divergence (for
example a `camPos.W` or directional-light seam) to fix, not a tolerance to widen.

## Testing Strategy

The test IS the deliverable: it is the regression lock. It fails loudly if either
shader drifts from the other, so a future edit to `FragmentShader` or
`kernels.Shade` that breaks the shared core is caught on all platforms (it does
not need a GPU). Verified passing locally; CI-pending.

## Out of scope (separate bounded specs)

- Multi-backend render (compile `ShadeSrc` to GLSL/SPIR-V in render; lift the
  Metal-only restriction). This is a forward feature, not unification.
- Shadow and AO kernels onto the author-once path.
- Ray tracing as a renderer mode.

## Deliverable

`render/shading_equiv_test.go` locking `shader.FragmentShader` to `kernels.Shade`
within 1 LSB. With it, the unification proper is complete: GPU by default with
CPU fallback, one shading authority, CPU and GPU proven to agree. The remaining
phases (multi-backend, ray tracing) are forward features for the engine, to be
prioritized with the user.
