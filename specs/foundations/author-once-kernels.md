---
title: Author-once kernels: gpumath + compiler method lowering, one source for CPU and GPU
status: implemented (CI-verified)
depends_on:
  - foundations/unified-renderer.md
  - foundations/gpu-phase2-goshader.md
affects:
  - gpu/shader/gpumath
  - gpu/shader
effort: medium
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Author-once kernels

## Overview

This is the first bounded slice of the unified renderer
([unified-renderer.md](unified-renderer.md)). It delivers the single mechanism on
which the whole unification rests: **a kernel authored once in Go that runs as
ordinary Go on the CPU and compiles to MSL/GLSL/SPIR-V for the GPU**, proven on
one real kernel by cross-backend parity. It does **not** touch the renderer, the
pass pipeline, GPU-by-default, or the CPU executor wiring; those are later
bounded specs. Scope here is just: the shared math library, the compiler support
to lower it, and one parity proof.

## Current State

- The Go->shader compiler (`gpu/shader`, [gpu-phase2-goshader.md](
  gpu-phase2-goshader.md)) accepts a restricted Go DSL that uses **operator
  overloading** on vector/matrix types (`lp - wpos`, `m * v`, `col * s`). That is
  not valid Go, so such a kernel cannot also run on the CPU as Go (the blocker
  noted in the unified-renderer spec).
- The compiler already maps a whitelist of lowercase free-function builtins
  (`normalize`, `dot`, `pow`, `clamp`, ...) and the `Vec2/3/4`/`Mat4` type names
  (`gpu/shader/compile.go`: `builtins`, `goToMSLType`), and handles one method
  call (`Texture2D.Sample`) in `call()`.
- The parity harness (`gpu/parity_*_test.go`) compares a GPU kernel against a
  hand-written Go reference (`cpuShade`). Today the reference duplicates the
  kernel logic by hand.

## The mechanism

A kernel expresses vector math through **methods and capitalized free functions**
from a shared `gpumath` package instead of operators, so the source is valid Go:

| Operator DSL (today) | Author-once Go (`gpumath`) | Compiler lowers to |
| --- | --- | --- |
| `a - b` (Vec4) | `a.Sub(b)` | `(a - b)` |
| `a + b` | `a.Add(b)` | `(a + b)` |
| `a * b` (componentwise) | `a.Mul(b)` | `(a * b)` |
| `a * s` (vec*scalar) | `a.Scale(s)` | `(a * s)` |
| `a / s` | `a.Div(s)` | `(a / s)` |
| `m * v` (Mat4*Vec4) | `m.MulV(v)` | `(m * v)` |
| `normalize(v)` | `Normalize(v)` | `normalize(v)` |
| `dot(a,b)` | `Dot(a,b)` | `dot(a, b)` |
| `pow(x,y)` | `Pow(x,y)` | `pow(x, y)` |

The kernel `import . "poly.red/gpu/shader/gpumath"` so `Vec4{...}`, `Normalize`,
`a.Sub(b)` read unqualified, identical for the Go compiler and the shader
compiler. On the CPU the methods/functions execute as written; on the GPU the
compiler lowers them.

### Components

1. **`gpu/shader/gpumath`** (new): float32 `Vec2/3/4`, column-major `Mat4`, the
   methods in the table, and capitalized free functions (`Normalize`, `Dot`,
   `Length`, `MulV`, `Clampf`, `Pow`, trig, `Sqrt`, `Floor`/`Round`/`Ceil`,
   `Minf`/`Maxf`, ...). Pure Go, unit-tested. `Clampf`/`Minf`/`Maxf`/`Absf` are
   named to avoid clashing with Go 1.21 builtins `min`/`max`/`clamp`; the compiler
   maps them to `clamp`/`min`/`max`/`abs`.

2. **`gpu/shader` compiler support** (extend `compile.go`, MSL identity preserved):
   - In `call()`, recognize method calls on vector/matrix-typed receivers and
     lower them: `.Add/.Sub/.Mul/.Scale/.Div` to the binary operator,
     `.Dot/.Length/.Normalize` to the builtin, `.MulV` to `(m * v)`. Use the
     existing `inferType` to confirm the receiver is a vec/mat type.
   - Add capitalized aliases (`Normalize`->`normalize`, `Dot`->`dot`,
     `Clampf`->`clamp`, `Pow`->`pow`, ...) to the `builtins` map so the gpumath
     free-function spelling lowers to the shader builtin.
   - No change to the operator path or any existing kernel: MSL output for the
     current operator-DSL kernels stays byte-identical (verified by the existing
     `TestCompile` corpus).

3. **One author-once proof kernel.** Re-express the parity shading kernel
   (`gpu/parity_shared_test.go` `shadingKernelSrc`) in gpumath form in a real Go
   file under a `kernels`-style test package: it runs as Go on the CPU, and its
   source compiles to GPU via `Compile`/`CompileGLSL`. The parity test asserts
   GPU-result == this-kernel-run-as-Go == the existing reference, on
   Metal/GL/Vulkan.

## Testing Strategy

- **gpumath unit tests** (offline): vector/matrix identities (normalize length 1,
  dot, MulV vs a known matrix-vector product).
- **Compiler lowering tests** (offline, pure Go): compile a gpumath-form kernel
  to MSL and GLSL and assert the lowered output matches the operator-DSL output
  for the same logic (e.g. `a.Sub(b)` produces the same `(a - b)` MSL as the
  operator form). Assert the existing `TestCompile` corpus is unchanged (MSL
  byte-identical).
- **Author-once parity** (CI, all backends): the proof kernel run as Go on the
  CPU == the same kernel compiled and run on Metal/GL/Vulkan == the reference,
  reusing the parity harness and its per-backend jobs.

## Out of scope (separate bounded specs)

- The `Pass` abstraction and pass-pipeline refactor.
- GPU-by-default device acquisition and `render.CPU()`.
- Replacing the CPU `shader/` Blinn-Phong / wiring kernels into `render`.
- Vertex/fragment or ray kernels.

## Deliverable

`gpumath` + the compiler lowering + one parity-proven author-once kernel, all CI
green. This unblocks the renderer-side work by proving a single kernel source
drives both CPU and GPU identically.
