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

## Status: cgo-free Vulkan compute works end-to-end (CI-verified)

The full cgo-free Vulkan **compute path is proven in CI** on Mesa lavapipe
(`gpu/vk*_linux_test.go`, `.github/workflows/vk-probe.yml`): via purego (no cgo)
`TestVulkanComputeDispatch` builds an instance, logical device, host-visible
storage buffers, a shader module (GLSL compiled to SPIR-V by glslang), descriptor
set, compute pipeline, command buffer, `vkCmdDispatch`, queue submit, and reads
the doubled result back, matching the CPU. About 14 Vulkan structs marshal
correctly through purego. So the hard question ("does cgo-free Vulkan compute
work?") is answered: yes. What remains is wiring it behind the `backend`
interface and a Go to SPIR-V path so kernels are authored in Go (today via
glslang).

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
2. **Done.** Logical device + compute queue + host-visible storage buffer memory
   roundtrip, cgo-free, green in CI (`gpu/vkdevice_linux_test.go`). The create-info
   struct layouts marshal correctly through purego, so the device/memory
   foundation is proven.
3. **Done (via glslang).** SPIR-V is produced by compiling the kernel's GLSL with
   glslang in CI; a Go to SPIR-V emitter remains the principled follow-up.
4. **Done.** Descriptor set + compute pipeline + command buffer + `vkCmdDispatch`
   + readback: `TestVulkanComputeDispatch` doubles a buffer and matches the CPU,
   green in CI.
5. **Done.** Wired behind the `backend` interface (`gpu/backend_vk.go`):
   `gpu.Open(WithDriver(DriverVulkan))` drives Vulkan like Metal/GL.
   `TestVulkanBackendCompute` runs an add kernel through the public Device API and
   matches the CPU, green in CI. The compute pipeline + descriptor set are built
   lazily from the recorded bindings at commit. Remaining: render pipeline; a
   Go to SPIR-V emitter; then Windows (`vulkan-1.dll`).
