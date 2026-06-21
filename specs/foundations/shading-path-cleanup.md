---
title: "Shading-path cleanup: remove dead code, de-drift kernel copies"
status: drafted
depends_on:
  - foundations/render-deferred-author-once.md
affects:
  - shader/blinn.go
  - shader/blinn_old.go
  - gpu/shader/validate_test.go
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Shading-path cleanup: remove dead code, de-drift kernel copies

## Overview

A maintenance slice that pays down debt the author-once-kernel migration left
behind (see [render-deferred-author-once.md](render-deferred-author-once.md) and
[render-shading-equivalence.md](render-shading-equivalence.md)). Two audits found
the shading path carries dead code and a misleading filename, and the compiler
test holds a verbatim kernel copy whose "copy of render/gpudeferred.go" comment
is now false (render's deferred kernel moved to `kernels.ShadeSrc`). This slice
is behavior-preserving: remove the dead type, fix the misleading name, and point
the test at the canonical source. No kernel logic changes.

## Current State

- `shader/blinn.go` defines `BlinnShader` (+ `Vertex`/`Fragment` methods and a
  `var _ Program = &BlinnShader{}` assertion). Verified DEAD: no reference exists
  outside blinn.go. `BasicShader` (shader/program.go) is live (raster tests,
  shader tests) and stays.
- `shader/blinn_old.go` holds `FragmentShader`, the LIVE CPU Blinn-Phong fragment
  shader (called at render/raster.go:290). The `_old` name inverts reality: it is
  current, not legacy. It is intentionally separate from the GPU `kernels.Shade`
  (per render-shading-equivalence.md they are locked equivalent, not merged).
- `gpu/shader/validate_test.go::TestCompileAcceptsRealKernels` feeds
  `deferredKernelSrc` (a Scene-uniform copy of the pre-migration deferred kernel)
  to assert the compiler accepts a realistic kernel. The comment claims it is a
  verbatim copy of render/gpudeferred.go, which no longer defines that kernel.
- `gpu/shader/gpumath/kernels` exposes `ShadeSrc`, the canonical deferred kernel
  source (storage-buffer scene), already compiled in CI to MSL/GLSL/SPIR-V.

## Components

### Remove dead `BlinnShader`

Delete `BlinnShader`, its `Vertex`/`Fragment` methods, and the
`var _ Program = &BlinnShader{}` assertion from `shader/blinn.go`. Keep the
`Program` interface and `BasicShader`. Confirm `go build ./...` and the shader
tests still pass.

### Rename `blinn_old.go` -> `blinn_cpu.go`

Rename the file (git mv) so the name reflects that `FragmentShader` is the live
CPU fragment shader. Add a short doc comment on `FragmentShader` noting it is the
CPU deferred shader, locked equivalent to `kernels.Shade`
(render-shading-equivalence.md). No code change.

### De-drift the compiler-validation copy

In `gpu/shader/validate_test.go`, replace the local `deferredKernelSrc` constant
with `kernels.ShadeSrc` (import `poly.red/gpu/shader/gpumath/kernels`; no import
cycle, kernels does not import gpu/shader). The test still asserts the compiler
accepts the real deferred kernel, now from the canonical source, so it cannot
drift again. Remove the false "verbatim copy" comment.

## Out of scope (separate bounded specs)

- Migrating `shadowKernel` / `aoKernel` / `srgbKernel` (still inline DSL in
  render/) into author-once `kernels` package with Go + GPU co-authoring like
  `Shade`. Larger: needs CPU Go versions and parity. Own spec.
- De-drifting `gpu/backend_gl_deferred_linux_test.go` (an independent uniform-path
  GL test): folded into the kernel-migration slice, where the canonical shadow/AO
  sources become available.
- Documenting / reshaping the `material` global pool.

## Testing Strategy

Behavior-preserving, so the existing suites are the guard: `go build ./...`,
`go test ./shader/... ./gpu/shader/... ./render/...` stay green on darwin; the
compiler-validation test now compiles `kernels.ShadeSrc`. No new behavior to
test beyond what the move exercises.

## Deliverable

Dead `BlinnShader` gone; `blinn_old.go` renamed to `blinn_cpu.go` with a clarifying
doc; `validate_test.go` validating the canonical `kernels.ShadeSrc` instead of a
drifted copy. Zero behavior change; the shading path reads honestly.
