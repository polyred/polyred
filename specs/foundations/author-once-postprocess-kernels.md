---
title: "Author-once shadow / AO / sRGB kernels"
status: drafted
depends_on:
  - foundations/render-deferred-author-once.md
  - foundations/render-multibackend-kernels.md
affects:
  - gpu/shader/gpumath/kernels
  - render/gpudeferred.go
  - render/gpugamma.go
effort: medium
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Author-once shadow / AO / sRGB kernels

## Overview

Finishes the author-once-kernel migration. `kernels.Shade` is the single source
for deferred shading, but the renderer's other three GPU kernels still live as
inline DSL strings in `render/`: `shadowKernel`, `aoKernel`
(render/gpudeferred.go) and `srgbKernel` (render/gpugamma.go). This slice moves
all three into `gpu/shader/gpumath/kernels` as author-once Go (runs as Go on the
CPU, compiles to MSL/GLSL on the GPU), mirroring `Shade`. After it, no kernel DSL
string remains in `render/` and the compiler-test copies can reference the
canonical exported sources.

## Current State

- `kernels` holds `Shade`/`ShadeSrc` (shade.go + embed.go); the pattern is: Go
  using `import . "poly.red/gpu/shader/gpumath"`, source embedded via `//go:embed`,
  compiled by render through `kernelModule` (MSL for Metal, GLSL for GL).
- `render/gpudeferred.go` `shadowKernel` (`Shadow`) and `aoKernel` (`AO`), and
  `render/gpugamma.go` `srgbKernel` (`SRGB`), are operator-form DSL strings, each
  taking a struct uniform (`ShadowU` / `AOU`) bound via `uniformBuf`. `runShadowKernel`,
  `runAOKernel`, `runGamma` compile and dispatch them.
- gpumath provides every builtin these need (`Pow`, `Floor`, `Round`, `Clampf`,
  `Cos`, `Sin`, `Sqrt`, `Atan`, `Maxf`) and `Mat4`/`M4`/`Mat4.MulV`.
- The GPU shadow/AO/gamma output is already covered: render darwin tests render
  scenes with shadows/AO/gamma on the GPU and compare to the CPU default.

## Components

### `kernels/srgb.go`, `kernels/shadow.go`, `kernels/ao.go` (new)

Each re-expresses its kernel as author-once Go using gpumath:
- `SRGB(gid uint, in, out []float32)` — scalar transfer using `Pow`.
- `Shadow(gid uint, fragxyz, recv, depths, mats, color, su []float32)` — uses
  `M4`/`Mat4.MulV` for the light-clip transform, `Pow`/`Floor`/`Clampf`/`Round`.
- `AO(gid uint, fragxyz, aoflag, depthbuf, color, au []float32)` — uses `Cos`,
  `Sin`, `Sqrt`, `Atan`, `Maxf`, `Pow`, `Floor`, `Clampf`, `Round`.

The struct uniforms (`ShadowU`, `AOU`) become trailing `[]float32` storage
buffers (`su`, `au`), matching `Shade`'s `scene []float32`. The packed values are
identical (`su = [W, DepthLen, N, _]`, `au = [W, H, _, _]`), so GPU output is
unchanged; this also removes render's last uniform buffers.

### `kernels/embed.go`

Add `//go:embed srgb.go` -> `SRGBSrc`, `shadow.go` -> `ShadowSrc`, `ao.go` ->
`AOSrc`, alongside `ShadeSrc`.

### render call sites

`runGamma`, `runShadowKernel`, `runAOKernel` compile `kernels.SRGBSrc /
ShadowSrc / AOSrc` via `kernelModule` (already backend-aware) and bind the
former uniform as a storage buffer (`storageBuf` instead of `uniformBuf`). Delete
the three inline consts and the now-unused `uniformBuf` helper.

### Test-copy consolidation

`gpu/shader/validate_test.go`'s corpus `shadow`/`ao` entries reference
`kernels.ShadowSrc` / `kernels.AOSrc` instead of local copies (remove
`shadowKernelSrc` / `aoKernelSrc`). `glsl_test.go`'s `uniformSceneKernelSrc` stays
(it is the deliberate UBO fixture, unrelated).

## Testing Strategy

- `kernels` unit tests: call `SRGB`/`Shadow`/`AO` as Go on a tiny controlled
  input and assert the expected output (proves the migrated source is valid Go
  and behaves, the CPU half of author-once). For SRGB, check the transfer curve at
  a couple of points; for Shadow/AO, a single fragment with hand-set buffers.
- Existing render darwin GPU tests (shadow / AO / gamma vs CPU) stay green: they
  are the guard that the uniform-to-storage rebind and the gpumath rewrite did not
  change GPU output.
- `gpu/shader` compiler tests still compile the (now canonical) shadow/AO sources.
- gl-probe: the multi-backend kernels still compile to GLSL (kernelModule path).

## Out of scope

- Equivalence-locking each kernel's run-as-Go against the CPU default impl
  (render shadow uses shadingVisibility, AO uses material/ao.go, gamma uses a LUT).
  Those CPU defaults stay; the run-as-Go here is a behavioral unit check, not a
  full 1-LSB lock like shading-equivalence. A lock per kernel can be a later slice.
- The `material` global pool cleanup (separate slice, next).

## Deliverable

Three new author-once kernels in `kernels/` with embedded sources; render compiles
them via `kernelModule`; the inline DSL strings and `uniformBuf` gone from render;
compiler-test copies point at the canonical sources; unit tests for the run-as-Go
path; existing GPU-vs-CPU tests green. The migration to author-once kernels is
complete: no shading kernel source lives in `render/` anymore.
