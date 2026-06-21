---
title: cgo-free Vulkan compute backend for the GPU abstraction
status: drafted (viability proven)
depends_on:
  - foundations/gpu-gl-backend.md
affects:
  - gpu/backend_vk.go (new)
  - gpu/shader/compile.go
  - gpu/vkprobe_linux_test.go
effort: xlarge
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# cgo-free Vulkan compute backend

## Overview

A third driver behind the same private `backend` interface: a cgo-free Vulkan
compute backend, after Metal and GL. Vulkan is the most portable modern API
(Linux/Windows/Android, and macOS via MoltenVK), so it widens reach the most.

## Status: viability proven, backend not yet built

The cgo-free Vulkan path is **proven in CI** (`gpu/vkprobe_linux_test.go`,
`.github/workflows/vk-probe.yml`): via purego (no cgo) the probe creates a Vulkan
instance, enumerates physical devices, and confirms a compute-capable queue
family. On the stock `ubuntu-latest` runner this is served by Mesa's **lavapipe**
software ICD (`llvmpipe`, device type CPU), so the full backend will be
CI-verifiable in software, exactly as the GL backend is. This de-risks the work;
the backend itself is the large remaining piece.

## The hard part: shader input is SPIR-V, not text

Unlike Metal (MSL) and GL (GLSL), Vulkan consumes **SPIR-V** binary modules. So
this backend needs one of:
- a Go to SPIR-V emitter in `gpu/shader` (the principled path; large), or
- compiling the existing GLSL (`CompileGLSL`) to SPIR-V offline and embedding the
  result (avoids a runtime dependency but adds a build step), or
- a runtime GLSL to SPIR-V step (needs glslang/shaderc, a heavy dependency,
  arguably against the cgo-free/lean spirit).

The first or second is preferred; this is the main design decision to settle
before implementation.

## Components (sketch)

- `gpu/backend_vk.go` (`//go:build linux`, later windows): instance/device/queue
  setup (probe code is the seed), `VkDeviceMemory` + `VkBuffer` storage buffers,
  a `VkShaderModule` from SPIR-V, a compute `VkPipeline` + `VkPipelineLayout` +
  `VkDescriptorSet`, a command pool/buffer, `vkCmdDispatch`, and host-visible
  memory map for readback. All struct marshaling through purego (the probe shows
  the pattern: C-layout Go structs, pointers via `unsafe.Pointer`).
- `gpu/shader`: the SPIR-V path above.

## Testing Strategy

- **Conformance (CI, software):** reuse the backend-agnostic compute conformance
  (matrix Add/Sub/Sqrt, the deferred kernel) through the Device API on lavapipe,
  gated on `POLYRED_VK_PROBE`-style env, mirroring the GL job. Already proven that
  the runner provides a compute-capable Vulkan device.
- **Build gate:** `GOOS=linux CGO_ENABLED=0 go build ./gpu/...`.

## Sequencing

1. **Done.** Viability probe (instance + compute device), green in CI.
2. Settle the SPIR-V story (Go to SPIR-V vs offline-compiled GLSL).
3. Device/queue/buffer/pipeline/dispatch/readback for compute; conformance.
4. Render pipeline; then Windows (same purego loader, `vulkan-1.dll`).
