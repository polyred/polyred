# polyred specs

Design specs for non-trivial work, written **before** implementation. Each spec
captures the problem, current state, architecture, and testing strategy so the
implementation has a reviewable target.

Workflow: write/iterate a spec here → implement against it → diff the result
back. The high-level GPU architecture and locked decisions live in
[`docs/gpu-abstraction.md`](../docs/gpu-abstraction.md); per-phase implementation
specs live here.

## Tracks

- **foundations** — core abstraction interfaces the rest of the engine builds on
  (e.g. the GPU `Device` abstraction).

## foundations

| Spec | Status | Deliverable |
| --- | --- | --- |
| [gpu-phase1-foundation.md](foundations/gpu-phase1-foundation.md) | Not started | Fold `poly.red/x/gpu` into polyred, cgo-free Metal compute via purego, the `Device` API skeleton, and the matrix demo through it |
