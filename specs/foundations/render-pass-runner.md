---
title: Render pass runner: uniform GPU-or-CPU dispatch for a pass
status: implemented (CI-verified)
depends_on:
  - foundations/unified-renderer.md
affects:
  - render
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Render pass runner

## Overview

The second bounded slice of the unified renderer ([unified-renderer.md](
unified-renderer.md)). It introduces one small seam: a `runPass` helper that
takes a pass's GPU function and CPU function and runs the GPU one when a device
is present (falling back to CPU on any error), recording which path ran. It
refactors the renderer's two existing ad-hoc offloads (deferred shading, gamma)
onto this seam. **Behavior-preserving**: no visual change, no public API change,
no new GPU passes. This makes "a pass runs on GPU-if-available else CPU" a single
reusable mechanism and makes the chosen path observable per pass, which the later
GPU-by-default and new-GPU-pass specs build on.

## Current State

`render/raster.go` has the same try-GPU-then-CPU shape inlined twice:

- `passDeferred` (lines ~185-228): `if r.cfg.GPUDevice != nil { if err :=
  gpuDeferredShade(...); err == nil { gpuDeferredUsed = true; return } }` then the
  CPU path `r.DrawFragments(buf, r.shade)`.
- `passAntialiasing` (lines ~268-288): `if GammaCorrect { if GPUDevice != nil {
  if err := gpuGammaCorrect(...); err != nil { CPU } } else { CPU } }`.

The GPU-path-taken signal is a single package-level `gpuDeferredUsed` bool (used
only by tests). There is no shared runner and no per-pass record.

## Components

### `render`: `runPass` + per-pass path record

- A small method `runPass(name string, gpu func() error, cpu func())`:
  - if `r.cfg.GPUDevice != nil` and `gpu != nil`: call `gpu()`; on `nil` error
    record path = GPU for `name` and return; on error, log (debug) and fall
    through.
  - call `cpu()`; record path = CPU for `name`.
- A `map[string]passPath` (or small struct) on the renderer recording the path
  per named pass for the last frame. `gpuDeferredUsed` is replaced by querying
  this record (`r.passRan("deferred") == pathGPU`); update the tests accordingly.
- Refactor `passDeferred` and `passAntialiasing` to call `runPass` with their
  existing GPU and CPU closures. The GPU closures wrap `gpuDeferredShade` /
  `gpuGammaCorrect`; the CPU closures wrap the existing `DrawFragments` calls.
  No logic moves; only the dispatch is centralized.

This is deliberately not a full `[]Pass` pipeline (that is a later spec); it is
the minimal extraction of the dispatch decision.

## Error Handling

Unchanged from today: a GPU error falls back to the CPU path for that pass. The
record captures which path actually executed so tests and (later) diagnostics can
assert it.

## Testing Strategy

- **Behavior-preserving**: the existing `render` tests (CPU and, on macOS, the GPU
  deferred/gamma parity tests) pass unchanged.
- Replace `gpuDeferredUsed`-based assertions with `r.passRan("deferred")`.
- Add a unit test that, with no device, `runPass` runs the CPU closure and records
  CPU; with a device and a GPU closure returning an error, it falls back to CPU
  and records CPU; with a device and a succeeding GPU closure, it records GPU.
  (The closure-level test needs no real GPU.)

## Out of scope (separate bounded specs)

- The full `[]Pass` pipeline / reorderable passes.
- GPU-by-default device acquisition (`gpu.Open()` in `NewRenderer`, `render.CPU()`).
- Wiring the author-once `kernels.Shade` into the deferred pass (CPU+GPU from one
  source): its own bounded spec on top of this and author-once-kernels.
- New GPU passes (forward raster, shadow, AO on GPU).

## Deliverable

`runPass` + a per-pass path record, with the two existing offloads refactored onto
it and tests updated. No behavior change; CI green (CPU everywhere, GPU paths on
macOS).
