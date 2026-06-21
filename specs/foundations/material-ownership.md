---
title: "Material ownership: de-globalize the material pool"
status: drafted (design, not scheduled)
depends_on: []
affects:
  - material/pool.go
  - render/raster.go
  - render/gpudeferred.go
  - geometry/primitive
  - scene
effort: large
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Material ownership: de-globalize the material pool

## Overview

This is a DESIGN spec, not scheduled work. It captures the real fix for the
"suboptimal abstraction" the cleanup audit flagged: `material/pool.go` is a
process-wide mutable registry (`Get`/`Put`/`Del` over a global map). Documenting
the contract (done) does not remove the smell; de-globalizing it is an
architecture change with real blast radius, so it is written up here for a
deliberate decision rather than folded into a tidy-up.

## Current State (verified)

- `material/pool.go`: a package-level `pool` (RWMutex + `idToMat`/`matToId` maps +
  incremental `allocID`). `Put` interns a material and returns an `ID`; `Get`
  resolves an `ID`; `Del` removes one (used by `model/load.go:151`). ID 0 is the
  default, seeded in `init`.
- `ID` is `int64`. Negative IDs are deliberately never pooled: `Get(neg)` returns
  nil, which render reads as "use vertex color". Producers: `model/plane.go`,
  `geometry/primitive/polygon.go` (`MaterialID: -1`). Consumers: `render/raster.go`
  (`shade`), `render/gpudeferred.go` (G-buffer marshaling).
- `primitive.Fragment.MaterialID` and `Triangle.MaterialID` carry the global ID
  through geometry.
- The GPU deferred path builds a SECOND, per-frame materials table keyed by
  `*BlinnPhong` (`matIndex` in gpudeferred.go), independent of the global ID. So
  two indexing schemes already coexist: global pool ID vs per-frame table index.

## Problems

1. Global mutable state: material IDs are shared across all renderers/scenes;
   there is no ownership boundary. Two scenes silently share the same ID space.
2. No lifecycle: materials are freed only by explicit `Del`; nothing scopes a
   material to the scene that uses it.
3. Two indexing schemes (global ID + GPU per-frame table) to keep in sync.

## Design directions (to choose between)

- **A. Scene-owned materials.** A `Scene` (or `Geometry`) owns its materials;
  `Fragment`/`Triangle` carry a material reference or a scene-local index instead
  of a global ID. The pool becomes per-scene. Largest change: every
  `material.Get` in render resolves against the scene; geometry loading attaches
  materials to the geometry, not a global pool. Removes the global entirely.
- **B. Keep the pool, add ownership handles.** Keep the registry but hand out a
  scoped handle (e.g. a `Set` a renderer owns) so IDs are not process-global;
  smaller change, but still a registry.
- **C. Unify the two indexing schemes only.** Make the GPU per-frame table the
  single material indexing path and derive the CPU path from it (or vice versa),
  without removing the global pool. Narrowest; addresses problem 3, not 1-2.

The negative-ID "vertex color" convention must be preserved (or replaced by an
explicit per-fragment flag) under any option.

## Blast radius

`render/raster.go`, `render/gpudeferred.go` (both indexing schemes),
`geometry/primitive` (Fragment/Triangle.MaterialID), `model/load.go`,
`scene`, and every test that constructs materials. This is why it is a design
decision, not a cleanup commit.

## Recommendation

Defer until the engine direction is set (it intersects how scenes own GPU
resources). When scheduled, break down per the chosen option (A is the clean
end-state but the largest; C is a safe first step that removes the
dual-indexing). Until then, `material/pool.go` is documented as global state with
this spec as the pointer.

## Out of scope

Everything until scheduled. No code change ships from this spec as written.
