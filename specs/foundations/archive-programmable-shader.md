---
title: "Archive the programmable-shader pipeline + the attribute-map varyings"
status: implemented (CI-verified)
depends_on:
  - foundations/render-deferred-author-once.md
affects:
  - shader
  - geometry/primitive
  - render
  - cmd/polywine
effort: medium
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Archive the programmable-shader pipeline + the attribute-map varyings

## Overview

Archive the 2022-era programmable-shader design, which the engine's fixed-function
deferred path + author-once kernels have superseded. Two layers:

1. The `AttrSmooth`/`AttrFlat` attribute-map varyings on `primitive.Vertex`/
   `Fragment` (the programmable pipeline's flexible varying mechanism).
2. The `shader.Program` interface and its implementations (`BasicShader`,
   `TextureShader`) driven via `Renderer.DrawPrimitives(buf, tris, prog.Vertex)` +
   `DrawFragments(buf, prog.Fragment)`.

The alternative is already in place: the renderer's typed `Fragment` fields
(Nor/Col/UV/WordPos) + fixed-function `shade()` + the author-once `kernels.Shade`
(deferred). This is an archival-with-alternative, not a deletion: the live render
path does not use either layer.

## Current State (verified)

- `AttrSmooth`/`AttrFlat` (`map[AttrName]any`) are allocated empty in `NewVertex`
  (vertex.go:43-44) and deep-copied in `Vertex.Copy`. NO live code populates them
  with attributes; the only live reader is `render/raster_primitive.go`
  `interpoVaryings`, gated on `len(v1.AttrSmooth) > 0`, so it never runs. The
  programmable shaders that consumed them (the deleted `BlinnShader`) are gone.
  `AttrName` + `AttrPosition/Normal/Color/UV` consts and `shader.MVPAttr` are part
  of this dead mechanism.
- `shader.Program` (Vertex/Fragment methods), `BasicShader`, `TextureShader` live
  in `shader/`. `Renderer.DrawPrimitives` applies a vertex-shader func;
  `DrawFragments` applies a fragment func. The LIVE renderer uses `r.draw` (not
  DrawPrimitives) and `DrawFragments(buf, r.shade)` internally. The only external
  users of the programmable pipeline: `cmd/polywine` (`TextureShader`), and the
  `shader`/`render` tests.
- `DrawFragments` itself is NOT archived: the renderer uses it internally for the
  fragment loop (passDeferred/antialias). Only the `Program` abstraction and
  `DrawPrimitives` (external vertex-shader entry) are archived.

## Components

### Part 1: remove the dead varying maps (safe, no live consumer)

- `geometry/primitive/vertex.go`: drop `AttrSmooth`/`AttrFlat` fields, their
  init in `NewVertex`, and the copy loops in `Copy`. Drop `AttrName` +
  `AttrPosition/Normal/Color/UV` consts (verify no other use).
- `geometry/primitive/fragment.go`: drop `AttrSmooth`/`AttrFlat` fields.
- `render/raster_primitive.go`: drop `interpoVaryings` and its gated call.
- `shader/mvp.go`: drop `MVPAttr` (keep the `MVP` struct, which the renderer uses).
- Saves two map allocations per vertex; the typed fields remain the varying path.

### Part 2: archive the Program pipeline (needs the polywine port)

- Port `cmd/polywine` from `TextureShader` + `DrawPrimitives`/`DrawFragments` to
  the modern `render.NewRenderer(Scene, Camera, ...).Render()` path: load the
  bunny with its texture as a material, render the scene (add a light to match the
  textured look). The orbit control updates the camera each frame via
  `r.Options(Camera(...))`.
- Remove `shader.Program`, `BasicShader`, `TextureShader`, and
  `Renderer.DrawPrimitives`. Update/retire `shader/shader_test.go` and
  `render/raster_primitive_test.go` (they exercise the programmable pipeline).

## Testing Strategy

- Part 1: `go build ./...` + full test suite green (the maps are unread, so
  behavior is unchanged). A CPU-render hash before/after (like the material
  refactor) confirms byte-identical output.
- Part 2: polywine builds and renders the bunny via the scene renderer (manual
  visual check, plus `go build`); the engine's existing render/golden tests
  unaffected. Removed tests are replaced by the renderer's own coverage.

## Out of scope

- Touching `DrawFragments` (live, internal).
- The GPU forward rasterizer arc (separate).

## Deliverable

The programmable-shader abstraction and its dead varying maps are gone; the live
fixed-function + author-once path is the single shading model; polywine runs on
the modern renderer. The suboptimal abstraction the audit flagged is archived,
with the alternative already in place.
