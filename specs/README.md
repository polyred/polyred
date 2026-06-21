# polyred specs

Design specs for non-trivial work, written **before** implementation. Each spec
captures the problem, current state, architecture, and testing strategy so the
implementation has a reviewable target.

Workflow: write/iterate a spec here → implement against it → diff the result
back. The high-level GPU architecture and locked decisions live in
[`docs/gpu-abstraction.md`](../docs/gpu-abstraction.md); per-phase implementation
specs live here.

## Tracks

- **foundations**: core abstraction interfaces the rest of the engine builds on
  (e.g. the GPU `Device` abstraction).

## Known issues

- **Windows runtime windowing is unverified.** The Windows present path was
  ported from the defunct immediate-mode GL API to the modern textured-quad GLES
  blit the Linux path uses (see
  [windows-present-port.md](foundations/windows-present-port.md)), so the module
  builds on Windows again (verified by `GOOS=windows go build ./...` and the
  Windows CI job). Actually displaying a window and pumping Win32 messages still
  needs a Windows desktop session to verify; that runtime check is deferred.

## foundations

| Spec | Status | Deliverable |
| --- | --- | --- |
| [gpu-phase1-foundation.md](foundations/gpu-phase1-foundation.md) | **Done** | cgo-free Metal compute via purego, the `Device` API, and the matrix demo through it |
| [gpu-phase2-goshader.md](foundations/gpu-phase2-goshader.md) | **Done** | Go→shader compiler (compute + vertex/fragment → MSL): varyings, uniforms, swizzle, vector/matrix math, texture sampling, trig |
| [gpu-phase3-render.md](foundations/gpu-phase3-render.md) | **Done** | Render pipelines + the renderer's full deferred pass offloaded to the GPU: lights, multi-material, shadow maps (N lights), ambient occlusion, gamma; CPU-parity verified |
| [windows-present-port.md](foundations/windows-present-port.md) | **Build done, runtime deferred** | Windows window present ported to the modern textured-quad GLES blit; builds on Windows, runtime needs a Windows desktop |
| [gpu-gl-backend.md](foundations/gpu-gl-backend.md) | **Compute + render done, CI-verified** | cgo-free GLES 3.1 backend behind the `backend` interface: compute (storage + UBO) and render-to-texture (FBO), driven through the Device API and verified on Mesa llvmpipe (software, surfaceless) in CI. Follow-ups: engine integration, Go-to-GLSL render shaders, Vulkan/DX12 |
| [gpu-windowed-present.md](foundations/gpu-windowed-present.md) | **Surface API done (headless), CI-verified** | backend-agnostic swapchain (`gpu/surface.go`): acquire/present/resize, render-through-swapchain verified headless on the GL backend. Remaining: on-screen attachment (needs a display) |
| [gpu-vulkan-backend.md](foundations/gpu-vulkan-backend.md) | **Compute dispatch works (CI-verified)** | cgo-free Vulkan compute end-to-end on Mesa lavapipe: shader module (SPIR-V via glslang) + descriptor set + pipeline + vkCmdDispatch + readback, matched to CPU. Remaining: wire behind the backend interface, Go-to-SPIR-V, then Windows/DX12 |

The GPU abstraction's Metal-backend phases are complete: the renderer's deferred
shading runs on the GPU, cgo-free, with shaders authored in Go. Remaining work is
**breadth, not depth**: additional backends (GL/Vulkan/DX12) and windowed
present, each gated on a Linux/Windows machine, an SDK, or a display rather than
on design (see the per-spec notes and the project memory for exact recipes).
