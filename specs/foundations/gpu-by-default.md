---
title: GPU by default: auto-acquire a device, fall back to CPU
status: drafted
depends_on:
  - foundations/render-pass-runner.md
affects:
  - render
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# GPU by default

## Overview

The third bounded slice of the unified renderer ([unified-renderer.md](
unified-renderer.md)). It flips the renderer's default from "CPU, opt-in GPU via
`render.GPU(dev)`" to "**GPU by default, automatic CPU fallback**", which is the
behavior the unified renderer promises. `NewRenderer` acquires a GPU device on
its own; a new `render.CPU()` option forces the CPU path (for the parity
reference and benchmarks). The public `Render()` surface is unchanged. Per-pass
fallback (from `runPass`, the previous slice) makes this safe: scenes a GPU pass
cannot handle still run on the CPU.

## Current State

- `NewRenderer` builds a renderer with `cfg.GPUDevice == nil`; the GPU paths only
  run if the caller passes `render.GPU(dev)`.
- `runPass(name, gpuFn, cpuFn)` already routes a pass to GPU-if-device-else-CPU
  and records the path (`render-pass-runner.md`).
- The GPU device lifetime is the caller's; the renderer does not own it.
- Tests that need a CPU reference call `NewRenderer(opts...)` (no device) and
  compare against `NewRenderer(append(opts, GPU(dev))...)`.

## Components

### `render.CPU()` option

`func CPU() Option` sets `cfg.forceCPU = true`. When set, `NewRenderer` does not
acquire a device and all passes run on the CPU.

### `NewRenderer` device acquisition

- If `cfg.GPUDevice == nil` and not `cfg.forceCPU`: call `gpu.Open()`. On success,
  set `cfg.GPUDevice` and mark the device renderer-owned. On error (no device,
  e.g. headless CI without a software driver), leave it nil (all-CPU). Never fail
  `NewRenderer` because of GPU acquisition.
- `render.GPU(dev)` still works and takes precedence (an explicitly supplied,
  caller-owned device; not closed by the renderer).
- The renderer closes a device it opened itself in `Close()` and the finalizer
  (alongside the existing `sched.Release()`); a caller-supplied device is left
  untouched.

### Test updates

Tests that rely on `NewRenderer(opts...)` being CPU (the parity references) add
`render.CPU()` so they stay CPU regardless of an available device. The
`passOnGPU("deferred")`-style assertions on the GPU renderers are unchanged.

## Data Flow / Behavior

Unchanged frame flow. The only change: with a usable device present, the deferred
and gamma passes now run on the GPU automatically (with CPU fallback for
unsupported scenes), instead of only when `render.GPU(dev)` was passed.

## Error Handling

`gpu.Open()` failure is non-fatal: the renderer runs all-CPU. Per-pass GPU errors
fall back to CPU via `runPass` (unchanged).

## Testing Strategy

- A test that, on a machine with a device, `NewRenderer(Scene, Camera, ...)` (no
  explicit device) runs the deferred pass on the GPU (`passOnGPU("deferred")`),
  and `NewRenderer(..., CPU())` runs it on the CPU. Skips the GPU assertion when
  no device is available.
- Existing parity/golden tests stay correct after adding `CPU()` to their
  reference renderers; CI green on macOS (device present) and Linux/Windows
  (no device in the main jobs ⇒ all-CPU, behavior unchanged).
- A device-leak check is unnecessary beyond confirming `Close()`/finalizer closes
  the owned device.

## Out of scope (separate bounded specs)

- Wiring the author-once `kernels.Shade` into the deferred pass.
- New GPU passes (forward raster, shadow, AO on GPU).
- A shared/process-wide device cache (each renderer opens its own for now).

## Deliverable

`render.CPU()` + auto device acquisition in `NewRenderer` + renderer-owned device
cleanup + reference tests pinned to CPU. GPU becomes the default with safe CPU
fallback; CI green.
