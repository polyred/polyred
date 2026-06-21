---
title: cgo-free DirectX 12 compute backend for the GPU abstraction
status: drafted (viability proven)
depends_on:
  - foundations/gpu-vulkan-backend.md
affects:
  - gpu/backend_dx12.go (new)
  - gpu/dxprobe_windows_test.go
effort: xlarge
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# cgo-free DirectX 12 compute backend

## Overview

A fourth driver behind the `backend` interface: a cgo-free DirectX 12 compute
backend for Windows, after Metal, GL, and Vulkan.

## Status: viability proven (not hardware-gated)

The cgo-free DX12 path is **proven in CI** (`gpu/dxprobe_windows_test.go`,
`.github/workflows/dx-probe.yml`): on the stock `windows-latest` runner (no GPU),
`D3D12CreateDevice` succeeds on the default adapter (the Microsoft Basic Render
Driver / WARP software rasterizer), reached cgo-free through the `syscall`
package (native on Windows; no purego needed). So like GL (Mesa llvmpipe) and
Vulkan (lavapipe), DX12 is CI-verifiable in software on a standard runner; it
does not need GPU hardware, only the Windows runner. The probe also carries a
WARP-via-DXGI fallback (a COM vtable call) for runners without a default adapter.

## The hard parts

- **COM.** D3D12/DXGI are COM: each object is a pointer to a vtable; methods are
  called as `obj->vtbl[index](obj, args...)`. In Go that is
  `*(*uintptr)(unsafe.Pointer(obj))` to get the vtable, then
  `syscall.SyscallN(vtbl[index], obj, args...)`. Vtable indices come from the
  header inheritance order (IUnknown first, then the interface chain). This is
  the main source of risk and must be transcribed carefully per interface.
- **Shaders are DXBC/DXIL.** Compile HLSL to DXBC at runtime via
  `D3DCompile` in `d3dcompiler_47.dll` (cgo-free via syscall), mirroring how the
  Vulkan backend uses glslang for SPIR-V. A Go to HLSL/DXIL emitter is a later
  follow-up.

## Components (sketch)

`gpu/backend_dx12.go` (`//go:build windows`): device (probe is the seed), a
compute command queue/allocator/list, committed UPLOAD/READBACK + DEFAULT buffer
resources with UAVs, a root signature + compute pipeline state from compiled
HLSL, a descriptor heap, `Dispatch`, and a fence for completion + readback. All
COM via syscall vtable calls.

## Testing Strategy

- **Conformance (CI, software):** reuse the backend-agnostic compute conformance
  (matrix Add/Sub/Sqrt) through the Device API on WARP, gated on
  `POLYRED_DX_PROBE`-style env, mirroring the GL/Vulkan jobs.
- **Build gate:** `GOOS=windows CGO_ENABLED=0 go build ./gpu/...`.

## Sequencing

1. **Done.** Device-creation probe (default adapter + WARP fallback), green in CI.
2. HLSL to DXBC via `D3DCompile`; command queue/allocator/list; UAV buffers.
3. Root signature + compute PSO + descriptor heap + `Dispatch` + fence +
   readback; conformance.
4. Wire behind the `backend` interface (`gpu.Open(DriverD3D12)`).
