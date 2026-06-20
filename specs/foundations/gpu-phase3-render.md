---
title: GPU Abstraction Phase 3 — Render slice (Metal, headless + windowed)
status: drafted
depends_on:
  - foundations/gpu-phase1-foundation.md
  - foundations/gpu-phase2-goshader.md
affects:
  - gpu/
  - gpu/mtl/
  - gpu/shader/
  - render/
effort: xlarge
created: 2026-06-20
updated: 2026-06-20
author: changkun
dispatched_task_id: null
---

# GPU Abstraction Phase 3 — Render slice

## Overview

Extend the `Device` API from compute-only to rendering: add render pipelines,
render passes, and textures as render targets, prove them with a headless
triangle render on Metal, then map polyred's deferred shading pass onto the GPU.
Supports both offscreen render-to-`*image.RGBA` (CI-testable) and windowed
present. Vertex/fragment kernels extend the Go→shader compiler (Phase 2).

## Current State

- Phase 1: `Device` API + cgo-free Metal **compute** backend (`gpu/device.go`,
  `gpu/backend_darwin.go`, `gpu/mtl`). `gpu/mtl` has Device/Queue/CommandBuffer/
  **compute** encoder/Buffer/Texture/blit — but **no render pipeline or render
  command encoder** yet.
- Phase 2: Go→MSL **compute** compiler (`gpu/shader`). No vertex/fragment profile.
- polyred renderer (`render/raster.go`): CPU passes shadow→forward→deferred→AA,
  `shader.Program` (Vertex/Fragment funcs), `MVP` uniforms, `FragmentBuffer`
  color+depth. `render.NewRenderer(opts).Render() *image.RGBA`.

## Components

### C1. Metal render plumbing (`gpu/mtl`)
Add cgo-free objc wrappers (same purego pattern as Phase 1): MTLRenderPipeline
Descriptor + state (`newRenderPipelineStateWithDescriptor:error:`),
MTLRenderPassDescriptor (color/depth attachments, load/clear/store),
`renderCommandEncoderWithDescriptor:`, set vertex/fragment buffers, `setVertexBytes`,
`drawPrimitives:vertexStart:vertexCount:` / indexed draw, vertex descriptor.

### C2. Device API render types (`gpu/device.go`)
`Texture`/`TextureDescriptor` (render-target + sampled), `RenderPipeline`/
`RenderPipelineDescriptor` (vertex+fragment modules, vertex layout, color/depth
formats), `RenderPassDescriptor` (attachments), `RenderPass`
(SetPipeline/SetBindGroup/SetVertexBuffer/SetIndexBuffer/Draw/DrawIndexed/End),
`CommandEncoder.BeginRenderPass`. Backend interface gains the render methods.

### C3. Go→shader vertex/fragment (`gpu/shader`)
Recognize vertex (`func(...) Vertex`) and fragment (`func(...) Color`) kernel
shapes; emit MSL `[[stage_in]]`/`[[position]]`/`[[color(0)]]`. Reuse the Phase 2
expression/statement translator; add the stage attributes and `float4`/`math.Vec`
types.

### C4. Headless render proof
Render a single clip-space triangle to an offscreen RGBA `Texture`, blit/readback
to `*image.RGBA`, assert interior pixels match the expected color. cgo-free.

### C5. Renderer integration (`render/`)
Add `render.Backend(CPU | GPU(dev))`. First GPU target: `passDeferred`
(`render/raster.go`) as a compute pass over G-buffer textures (most
self-contained). Map `shader.Program`→pipeline, `MVP`→uniform buffer,
`FragmentBuffer`→render-target textures. Exit: a scene renders within tolerance
on CPU and GPU.

### C6. Windowed present
Reuse `gpu/ctx/ca` (CAMetalLayer) for on-screen present via `PresentDrawable`.
Headless path stays the default for tests.

## Testing Strategy
- Headless triangle: render→readback→assert pixel colors (deterministic, CI).
- Parity: a known scene rendered via CPU vs GPU deferred pass, compare images
  within tolerance (reuse `internal/imageutil` diff).
- Go→shader vertex/fragment: golden MSL + real `NewShaderModule` compile.
- `CGO_ENABLED=0` build/test gate for the Metal render path.

## Progress

- **C1 Metal render plumbing — DONE** (`gpu/mtl/render_darwin.go`): render
  pipeline state, render pass descriptor, render command encoder, texture usage,
  `MTLClearColor` (4×float64 HFA struct-by-value), `MTLRegion` readback —
  cgo-free, spike-validated before implementation.
- **C2 Device API render types — DONE** (`gpu/render.go`): `Texture`/
  `TextureDescriptor`, `RenderPipeline`/`RenderPipelineDescriptor`,
  `RenderPassDescriptor`, `RenderPass` (SetPipeline/SetBindGroup/SetVertexBuffer/
  Draw/End), `CommandEncoder.BeginRenderPass`, `Texture.ReadPixels`.
- **C4 Headless render proof — DONE** (`gpu/render_darwin_test.go`): a triangle
  renders to an offscreen RGBA texture through the Device API and reads back red
  at center, cgo-free.
- **C3 Go→shader vertex/fragment — DONE** (`gpu/shader`, commit `3f56ffc`): `//gpu:vertex`/`//gpu:fragment` directives, Vec4→float4, value returns; a triangle rendered headless from Go-authored vertex+fragment shaders (`gpu/shader/render_darwin_test.go`).
- **C5 renderer integration (`passDeferred` on GPU) — TODO**.
- **C6 windowed present — TODO**.

## Notes
- Largest phase; will likely be broken into tasks (C1 render plumbing is the
  prerequisite and the bulk of new objc surface).
- GL render path + GLSL vertex/fragment needs the cgo-free GL backend and Linux
  to verify — separate phase.
