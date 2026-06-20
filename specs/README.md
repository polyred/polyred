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

## Known issues

- **Windows `app` windowing is broken (pre-existing).** `app/ctx_gl_windows.go`
  and `app/window_windows.go` call a defunct package-level immediate-mode GL API
  (`gl.MakeCurrent`, `gl.DrawBuffer`, `gl.RasterPos2d`, `gl.DrawPixels`,
  `gl.PixelZoom`, `gl.Viewport`) that the restructured GLES `gpu/gl` (methods on
  `*Functions`) no longer provides. The Linux present path was modernized to a
  textured-quad blit (`window_linux.go`: CreateTexture/TexImage2D/VertexAttrib/
  shader); Windows was never ported. Fix = port the Windows context + present to
  the modern GLES approach (needs a Windows machine to verify). Unrelated to the
  GPU abstraction. CI is green on macOS + Linux; Windows fails here.

## foundations

| Spec | Status | Deliverable |
| --- | --- | --- |
| [gpu-phase1-foundation.md](foundations/gpu-phase1-foundation.md) | Compute slice done | cgo-free Metal compute via purego, the `Device` API, and the matrix demo through it |
| [gpu-phase2-goshader.md](foundations/gpu-phase2-goshader.md) | Compute→MSL done | Go→shader compute compiler (Go kernels → MSL) feeding the `Device` API |
| [gpu-phase3-render.md](foundations/gpu-phase3-render.md) | Render plumbing + headless triangle done | Render slice: Metal render pipelines, headless triangle, deferred-pass GPU offload |
