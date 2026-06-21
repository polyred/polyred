---
title: "Multi-backend render kernels: compile per device driver"
status: drafted
depends_on:
  - foundations/render-deferred-author-once.md
  - foundations/gpu-gl-backend.md
affects:
  - render/gpudeferred.go
  - render/gpugamma.go
  - render/gpukernel.go
  - .github/workflows/gl-probe.yml
effort: medium
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Multi-backend render kernels: compile per device driver

## Overview

The first bounded slice of Phase 2 (multi-backend render) of the unified renderer
([unified-renderer.md](unified-renderer.md)). Today every render GPU pass hardcodes
Metal Shading Language: it calls `shader.Compile(src)` and binds
`gpu.ShaderSource{MSL: ks[entry].MSL}`. That is why `NewRenderer` auto-acquires
only a Metal device. This slice makes render's kernel compilation backend-aware:
one helper compiles a kernel for `dev.Driver()` (MSL for Metal, GLSL for GL), so
the render GPU passes run on a GL device too. It is CI-verified end to end on the
surfaceless Mesa probe.

This slice does NOT change `NewRenderer`'s automatic device acquisition (it stays
Metal-only). Flipping auto-acquisition to also try GL, and the device/surface
model behind it, is the next slice: opening GL on a standard CI runner (no
surfaceless Mesa) segfaults rather than erroring, so auto-GL needs care. Here, a
GL device reaches render only when a caller passes `render.GPU(glDev)` explicitly.

## Current State

- Four render GPU kernels each compile MSL-only (verified, gpudeferred.go /
  gpugamma.go):
  - deferred `Shade` (gpudeferred.go:360,364), now sourced from
    `kernels.ShadeSrc`.
  - `Shadow` (gpudeferred.go:432,436), `AO` (gpudeferred.go:496,500).
  - gamma `SRGB` (gpugamma.go:36,40).
- `gpu.Device.Driver()` reports the backend (`DriverMetal` / `DriverGL` /
  `DriverVulkan` / ...). `gpu.ShaderSource{MSL,GLSL,HLSL,SPIRV}`.
- `shader.Compile(src)` emits MSL; `shader.CompileGLSL(src)` emits GLSL (both pure
  Go, safe at runtime on every platform). SPIR-V needs an external glslang
  shell-out, so Vulkan render is deferred.
- The GL backend is proven through the public Device API and the author-once
  kernel: `gpu/parity_shared_test.go` runs `kernels.ShadeSrc` on GL == CPU oracle,
  and `gpu/backend_gl_linux_test.go` runs the GL conformance suite. Both gate on
  `EGL_PLATFORM=surfaceless` (set only in `.github/workflows/gl-probe.yml`) and
  skip otherwise, so standard `go test ./...` never opens GL (which would
  segfault on a runner without surfaceless Mesa).

## Components

### `render/gpukernel.go` (new)

`func kernelModule(dev *gpu.Device, src, entry string) (*gpu.ShaderModule, error)`:

- `dev.Driver()` -> `DriverMetal`: `shader.Compile(src)`, build
  `gpu.ShaderSource{MSL: ks[entry].MSL}`.
- `DriverGL`: `shader.CompileGLSL(src)`, build `gpu.ShaderSource{GLSL:
  ks[entry].GLSL}`.
- any other driver (Vulkan, DX12, Auto): return a sentinel
  `errKernelBackendUnsupported` so the caller's `runPass` falls back to CPU.
- create and return `dev.NewShaderModule(src)`.

### Swap the four compile sites

`runDeferredKernel`, `runShadowKernel`, `runAOKernel` (gpudeferred.go) and
`runGamma` (gpugamma.go) replace their `Compile` + `NewShaderModule(MSL)` pair
with a single `kernelModule(dev, src, entry)` call. Behavior on Metal is
unchanged (same MSL). On GL they now compile GLSL.

### No change to `NewRenderer`

Auto-acquisition still requests `DriverMetal` (gpu-by-default.md). Explicitly
passing `render.GPU(glDev)` is what exercises the GL path. Documented as such.

## Testing Strategy

- `render/gpukernel_test.go` (no device, all platforms): assert `kernelModule`
  selects GLSL for a GL-driver path and MSL for a Metal-driver path at the
  source-selection level, and returns `errKernelBackendUnsupported` for Vulkan.
  (Validate the ShaderSource chosen, not GPU execution; keep it device-free so it
  runs in standard CI.)
- `render/gl_render_linux_test.go` (`//go:build linux`, gated on
  `EGL_PLATFORM=surfaceless`, skip otherwise): open a GL device with
  `gpu.WithDriver(gpu.DriverGL)`, render a small synthetic scene (no file
  testdata, no shadow/AO) with `render.GPU(glDev)`, assert `passOnGPU("deferred")`
  and that the GPU image matches a `render.CPU()` render within the existing
  deferred tolerance. This is the end-to-end proof that render's deferred pass
  runs on GL.
- Extend `.github/workflows/gl-probe.yml`'s `-run` filter (or add a render step)
  to include the new render GL test, so it executes on the surfaceless Mesa
  runner. Standard `go test ./...` skips it (env unset) -> no segfault risk.
- Existing darwin deferred/gamma tests stay green (Metal path unchanged).

## Out of scope (separate bounded specs)

- Auto-acquiring a GL device in `NewRenderer` (auto-GL) and the device/surface
  model: the segfault-on-standard-CI constraint makes this its own slice.
- Vulkan render (runtime SPIR-V via glslang shell-out) and DX12 render.
- Sharing one device between the renderer and the windowed presenter (deferred
  from Phase 1; decide here once the device/surface model is set).

## Deliverable

`render/gpukernel.go` with `kernelModule`, the four render GPU passes routed
through it, a device-free unit test for source selection, and a CI-verified
render-deferred-on-GL test in the surfaceless probe. After this, render's GPU
passes are backend-agnostic and proven on GL; the Metal-only assumption is gone
from the kernel layer (only the auto-acquisition default remains Metal, by
design).
