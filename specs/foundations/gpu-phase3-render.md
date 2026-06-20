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
- **C5 renderer integration — gamma pass DONE** (commits `57c151b`, `45c9658`):
  `render.GPU(dev)` routes the renderer's gamma-correction pass through a GPU
  compute kernel (the engine's analytic sRGB, authored in Go) instead of the CPU
  fragment shader, with CPU fallback on error. A real bunny scene rendered with
  CPU vs GPU gamma is **bit-identical** (`render/gpugamma_test.go`, max channel
  diff 0). The renderer now genuinely consumes the `poly.red/gpu` abstraction.
  Also proven standalone: GPU diffuse lighting with a light **uniform** + normal
  **varying** (`gpu/shader/lightuniform_darwin_test.go`) — the structure of the
  deferred lighting pass.
- **C5 FULL deferred offload — DONE** (commit `e6368ea`): `passDeferred` runs the
  deferred Blinn-Phong shading on the GPU when `render.GPU(dev)` is set and the
  scene is supported (point lights + ambient, single Blinn-Phong material, no
  shadow map / AO): the per-fragment G-buffer (normal/world-pos/base-color) is
  marshaled, the proven Blinn-Phong kernel runs over all fragments, shaded
  colours are written back; CPU fallback otherwise. A 150×150 bunny scene renders
  **bit-identically** on CPU vs GPU (`render/gpudeferred_test.go`, max channel
  diff 0 over 90000 bytes, GPU path confirmed exercised).
  - **Directional lights** (commit `502ba5d`): kernel + marshaling handle point
    AND directional lights (`render/gpudeferred_directional_test.go`, parity ±1).
  - **Multiple materials** (commit `d54eb53`): per-fragment material index +
    deduplicated materials table; no-material fragments pass through as
    `info.Col`. A bunny+gopher scene renders bit-identically CPU vs GPU under
    `Workers(1)` (`render/gpudeferred_multimat_test.go`). Includes a gated
    pure-Go kernel self-check. NOTE: the concurrent forward pass is
    non-deterministic for overlapping objects, so multi-object parity must test
    single-worker.
  - **Shadow maps** (commit `f4f50c7`): a second compute pass
    (`render/gpudeferred.go` shadowKernel) projects each fragment to light space
    (combined `Viewport·lightProj·lightView·ScreenToWorld`, Mat4·vector), looks
    up the indexed shadow depth map, and darkens occluded fragments by
    `pow(0.5, occluded)` with the engine's exact round/clamp/truncate. Single
    casting light + `ReceiveShadow` materials; CPU fallback otherwise. A
    bunny+ground shadow scene renders CPU vs GPU within ±1
    (`render/gpudeferred_shadow_test.go`, `Workers(1)`).
  - **Multiple shadow-casting lights** (commit `97a3730`): the shadow kernel
    loops over N lights (per-light `float4x4` built in-kernel, packed depth
    maps). A 2-light bunny+ground scene matches CPU within ±1.
  - **Ambient occlusion** (commit `fe9c27f`): a final SSAO compute pass
    (`aoKernel`) mirrors `material/ao.go` — 8-direction depth-buffer march,
    `atan` elevation angles, `pow(total, 10000)`. The engine's `pow(.,10000)`
    amplifies GPU/CPU float differences, so a few contour-edge pixels diverge
    (0.74% of channels >8) while 99.26% match closely; exact parity is
    mathematically infeasible for this algorithm (the engine flags its own as
    "naive and super slow"). SSAO renders correctly on the GPU.
  - **The renderer's deferred pass is now fully offloaded:** point + directional
    lights + ambient + multiple materials + shadow maps (single & multi-light) +
    ambient occlusion + gamma, all on the cgo-free `poly.red/gpu` abstraction
    with CPU-parity (AO close-but-not-bit-identical by algorithm design).
- **C6 windowed present — TODO** (needs CAMetalLayer via `gpu/ctx/ca`, cgo).

## Notes
- Largest phase; will likely be broken into tasks (C1 render plumbing is the
  prerequisite and the bulk of new objc surface).
- GL render path + GLSL vertex/fragment needs the cgo-free GL backend and Linux
  to verify — separate phase.
