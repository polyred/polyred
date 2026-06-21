---
title: "GPU render pipeline: depth buffer (forward-rasterizer brick 1)"
status: implemented (CI-verified)
depends_on:
  - foundations/gpu-abstraction.md
affects:
  - gpu/render.go
  - gpu/backend.go
  - gpu/backend_darwin.go
  - gpu/mtl
effort: medium
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# GPU render pipeline: depth buffer (forward-rasterizer brick 1)

## Context: the GPU forward rasterizer is a multi-session arc

The engine-readiness frontier is moving the forward rasterizer (vertex transform +
triangle rasterization) from the CPU (`render` `passForward`/`draw`/`drawClipped`)
onto the GPU, feeding the existing GPU deferred shading with no CPU round-trip.
That needs three missing pillars in the `gpu` render pipeline, built brick by
brick, Metal-first (the established 2a pattern; GL/Vulkan follow):

1. **Depth buffer** (this spec): occlusion. Without it a real mesh renders wrong
   (painter-order), so it is the prerequisite for the first mesh-vs-CPU milestone.
2. **MRT / multiple color attachments**: a G-buffer (normals, worldpos, basecol)
   needs several render targets; the pass descriptor is single-attachment today.
3. **Scene wiring**: MVP vertex shader + mesh/camera upload + compare to the CPU
   forward pass.

This spec is brick 1 only.

## Current State

- `gpu/render.go`: `RenderPipelineDescriptor` has a single `ColorFormat`;
  `RenderPassDescriptor` has a single `ColorTexture` + `Load`/`ClearColor`. No
  depth. `RenderPipeline`/`RenderPass` (Draw, SetVertexBuffer, SetBindGroup).
- `gpu/backend.go`: `newRenderPipeline(vmod, ventry, fmod, fentry, color)` and
  `newTexture(format, w, h, renderTarget)`; the Metal/GL/Vulkan backends implement
  it. `gpu/mtl` has NO depth-stencil bindings.
- `cmd/gpudemo` renders a single triangle (no overlap, so no depth needed yet).
- `gpu/backend_gl_render_linux_test.go` renders a full-screen triangle to a texture
  (the headless render path), also no depth.

## Components

### `gpu/mtl` depth-stencil FFI (new)

Add the minimal Metal bindings (purego, following the existing selector pattern):
`MTLDepthStencilDescriptor` (+ `depthCompareFunction`, `depthWriteEnabled`),
`device.MakeDepthStencilState`, a depth `PixelFormatDepth32Float`, and the render
command encoder's `setDepthStencilState`. The render pass needs a
`MTLRenderPassDepthAttachmentDescriptor` (texture, loadAction clear, clearDepth,
storeAction).

### `gpu` API (depth in the descriptors)

- `RenderPipelineDescriptor.DepthFormat TextureFormat` (zero value = no depth, so
  existing callers are unchanged).
- `RenderPassDescriptor.DepthTexture *Texture` + `ClearDepth float64`.
- A depth `TextureFormat` value (e.g. `FormatDepth32Float`) and `newTexture`
  support for it as a render target.
- The `backend` interface signatures grow a depth format / depth attachment;
  GL and Vulkan return `ErrUnsupported` (or ignore) for now, Metal implements.

### Metal backend (`gpu/backend_darwin.go`)

`newRenderPipeline` sets the pipeline's depth attachment format and builds a
depth-stencil state (compare `less`, write on) when a depth format is given;
`beginRender` attaches the depth texture (clear to `ClearDepth`, default 1.0) and
the encoder binds the depth-stencil state. No depth format -> today's behavior.

## Testing Strategy

- `gpu` depth-correctness test (darwin, Metal): render two overlapping triangles
  into a color+depth target, a NEAR one and a FAR one drawn in both orders;
  read back pixels and assert the near triangle's color wins in the overlap region
  regardless of draw order. This is the occlusion proof depth exists to provide,
  and it fails (draw-order-dependent) without the depth attachment.
- Existing render tests (gpudemo, GL render) unchanged (depth format zero-valued).

## Out of scope (later bricks / specs)

- MRT / G-buffer attachments (brick 2).
- Scene/camera/mesh wiring and the MVP vertex shader (brick 3).
- GL and Vulkan depth (follow-up, like the 2a GL follow-up).

## Deliverable

A depth-capable GPU render pipeline on Metal, with a depth-correctness test. The
first occlusion-correct GPU render is possible; brick 2 (MRT) and brick 3 (scene
wiring) build the forward rasterizer on top.
