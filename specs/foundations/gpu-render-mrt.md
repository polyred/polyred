---
title: "GPU render pipeline: multiple render targets (forward-rasterizer brick 2)"
status: drafted
depends_on:
  - foundations/gpu-render-depth.md
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

# GPU render pipeline: multiple render targets (forward-rasterizer brick 2)

## Context

Brick 2 of the GPU forward rasterizer arc (depth -> MRT -> scene wiring; see
[gpu-render-depth.md](gpu-render-depth.md)). The deferred G-buffer needs several
color render targets in one pass (normals, world position, base color, ...), but
the render pipeline / pass support a single color attachment today. This brick
adds multiple color attachments, Metal-first (GL/Vulkan follow, like brick 1).

## Current State

- `gpu/render.go`: `RenderPipelineDescriptor.ColorFormat` (one);
  `RenderPassDescriptor.ColorTexture` + `ClearColor` (one). Depth was added in
  brick 1 (`DepthFormat`, `DepthTexture`, `ClearDepth`).
- `gpu/backend.go`: `newRenderPipeline(..., color, depth TextureFormat)`;
  `renderPassInfo{color, load, clearColor, depth, clearDepth}`.
- `gpu/mtl/render_darwin.go`: `RenderPipelineDescriptor.ColorPixelFormat`;
  `RenderPassDescriptor.ColorAttachment0` (sets `colorAttachments[0]`).
- A Metal fragment shader writes multiple targets by returning a struct with
  `[[color(0)]]`, `[[color(1)]]`, ... members.

## Components

Keep the single-attachment fields as attachment 0 (existing callers unchanged)
and add the extra attachments beyond it.

### `gpu` API

- `RenderPipelineDescriptor.ExtraColorFormats []TextureFormat` (attachments
  1..N). When empty, behavior is exactly today's single attachment.
- `RenderPassDescriptor.ExtraColorTargets []ColorTarget`, where
  `ColorTarget{ Texture *Texture; ClearColor [4]float64 }` (attachments 1..N,
  each cleared like attachment 0 when `Load == LoadClear`).
- `backend` interface: `newRenderPipeline` grows `extraColor []TextureFormat`;
  `renderPassInfo` grows `extraColor []renderColorTarget{tex backendTexture;
  clear [4]float64}`.

### `gpu/mtl`

- `RenderPipelineDescriptor.ExtraColorPixelFormats []PixelFormat`;
  `MakeRenderPipelineState` sets `colorAttachments[i].pixelFormat` for each.
- `RenderPassDescriptor.ExtraColorAttachments []ColorAttachment`; `objc()` sets
  `colorAttachments[i]` (texture, load, store, clear) for each.

### Metal backend

`newRenderPipeline` maps the extra gpu formats to mtl pixel formats;
`beginRender` builds the extra mtl color attachments from `renderPassInfo`. No
extras -> today's single-attachment path.

## Testing Strategy

- `gpu` MRT test (darwin, Metal): a full-screen triangle whose fragment shader
  writes a DISTINCT constant color to each of (say) 3 attachments; render into 3
  color targets in one pass; read each back and assert it holds its own color
  (target 1 != target 0 != target 2). Discriminating: a single-target pipeline
  cannot satisfy it; cross-checks that attachments are not aliased.
- Existing single-attachment tests (triangle, depth, gpudemo) unchanged.

## Out of scope

- Scene/camera/mesh wiring + MVP vertex shader (brick 3, uses depth + MRT).
- GL and Vulkan MRT (follow-up).
- Wiring the renderer's deferred G-buffer onto this (later, with brick 3).

## Deliverable

A render pipeline + pass that drive multiple color attachments on Metal, with an
MRT correctness test. The G-buffer the forward rasterizer needs is now
expressible; brick 3 wires a scene through depth + MRT.
