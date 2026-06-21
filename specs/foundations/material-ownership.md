---
title: "Material ownership: de-globalize the material pool"
status: implemented (CI-pending; Option A, staged, index-based)
depends_on: []
affects:
  - material/pool.go
  - render/raster.go
  - render/gpudeferred.go
  - geometry/primitive
  - scene
  - model
effort: large
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Material ownership: de-globalize the material pool

## Overview

The real fix for the "suboptimal abstraction" the cleanup audit flagged:
`material/pool.go` is a process-wide mutable registry (`Get`/`Put`/`Del` over a
global map). The user chose Option A (scene-owned materials, remove the global
pool) with full knowledge of the corrected scope.

## Chosen approach (decided 2026-06-21): index, staged

- **Index, not pointer.** A `material.Material` field on `primitive.Triangle`/
  `Fragment` is IMPOSSIBLE: `material` already depends on `primitive` (via
  `buffer` -- `material.Standard` holds `*buffer.Texture`; `material/ao.go` reads
  `MaterialID`), so a primitive -> material field is an import cycle. Primitives
  keep an integer; it just becomes scene-scoped instead of process-global.
- **Staged C -> A** (each slice golden/behavior-gated, committed separately):
  - **Slice 1 (C, behavior-preserving):** one material-resolution path. Today the
    CPU path is `material.Get(ID)` and the GPU path resolves `Get(ID)` then keys a
    per-frame table by `*BlinnPhong`. Introduce a single `resolveMaterial(id)`
    used by CPU shade (render/raster.go), GPU marshaling (render/gpudeferred.go),
    and `material/ao.go`. Still pool-backed. No output change.
  - **Slice 2 (A):** lift the registry from the global `pool` to scene ownership;
    rework `model.Load` for the load-before-scene ordering; delete the global pool.

## Verification (non-negotiable)

- The negative-ID -> vertex-color path (`-1` from model/plane.go, geometry/
  primitive/polygon.go; consumed via resolution -> nil) is pinned by a test BEFORE
  resolution is touched: `resolveMaterial(-1) == nil`, `resolveMaterial(0) ==
  default`, and a render smoke that a `-1` plane shows vertex color, not the
  default material.
- The repo has NO byte-exact render golden (TestRender only writes), and a
  byte-exact golden would be cross-platform flaky (rasterization float diffs).
  Parity is guarded by: the resolution unit tests, the existing GPU-vs-CPU
  multi-material test, and logical per-region pixel assertions, not a golden image.

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
