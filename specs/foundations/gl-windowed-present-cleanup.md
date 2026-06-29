---
title: "GL windowed present via the Device API; retire gpu/gl + gpu/ctx/egl"
status: in progress (brick 1)
depends_on:
  - foundations/gpu-windowed-present.md
  - foundations/cgo-free-windowed-present.md
affects:
  - gpu/backend.go
  - gpu/surface.go
  - gpu/backend_gl.go
  - gpu/backend_darwin.go
  - gpu/backend_vk.go
  - gpu/backend_other.go
  - app/window_linux.go
  - app/window_windows.go
  - app/ctx_egl_linux.go
  - app/ctx_gl_windows.go
  - gpu/gl
  - gpu/ctx/egl
created: 2026-06-29
updated: 2026-06-29
author: changkun
dispatched_task_id: null
---

# GL windowed present via the Device API; retire gpu/gl + gpu/ctx/egl

## Why

After the cgo-free port (`cgo-free-windowed-present.md`), the whole repo is
cgo-free, but windowed present is fragmented: linux/windows blit a CPU
`*image.RGBA` through the standalone `gpu/gl` (the `Functions` type) +
`gpu/ctx/egl`, while the GPU compute/render backend (`gpu/backend_gl.go`) has its
OWN purego EGL/GLES bindings and never touches those packages. So there are two
independent GL stacks. The goal: route windowed present through the GPU Device
API (one `backendSurface` seam), then delete the duplicate `gpu/gl` +
`gpu/ctx/egl`.

Map (2026-06-29): NO platform uses the Device/Surface API on-screen today.
Darwin blits the CPU image to a `CAMetalLayer` drawable via `gpu/mtl`+`gpu/ctx/ca`
(NOT `gpu/gl`). Linux/windows blit via `gpu/gl`+`gpu/ctx/egl` (textured quad +
`eglSwapBuffers`). `gpu/surface.go` is headless only (rotates textures + WaitIdle).
`gpu/backend_gl.go` is `//go:build linux` and surfaceless/pbuffer only.

## Design: the backendSurface seam (not a bespoke present path)

Extend the EXISTING `gpu.Surface` with an on-screen mode rather than adding a
parallel type, so Metal/VK/DX can join the same seam later (darwin's
`NextDrawable -> blit -> PresentDrawable` already fits it).

- `gpu/backend.go`: add to the `backend` interface
  `newWindowSurface(display, window uintptr, w, h int) (backendWindowSurface, error)`
  and the interface
  `backendWindowSurface { acquire() backendTexture; present() error; resize(w, h int) error; release() }`.
- `gpu/surface.go`: add `Device.CreateWindowSurface(WindowSurfaceDescriptor)`
  (native display+window handles), an on-screen `Surface` (holds a
  `backendWindowSurface`), and a convenience `Surface.PresentImage(img *image.RGBA)`
  = `acquire -> texture.write(img) -> present`. `AcquireNextTexture`/`Present`/
  `Resize` dispatch to the backend surface when on-screen.
- GL impl (`gpu/backend_gl.go`): `eglWindowBit` added to the config; a window
  surface created with `eglCreateWindowSurface(display-from-the-app, window)`;
  `acquire()` returns a persistent FBO-backed render-target texture; `present()`
  blits that texture's FBO to the default framebuffer (0) and `eglSwapBuffers`;
  `resize` reallocates. Needs new symbols: `eglCreateWindowSurface`,
  `eglDestroySurface`, `eglSwapBuffers`, `glBlitFramebuffer`.
- Other backends (`backend_darwin.go`, `backend_vk.go`, `backend_other.go`):
  `newWindowSurface` returns `ErrUnsupported` for now (interface compliance;
  darwin keeps its working Metal blit — out of scope here, future seam target).

### Thread / context ownership (the real runtime risk)

`backend_gl.go` marshals ALL GL onto one locked OS thread via `do()`. EGL's
current context is per-thread, so the window surface MUST be current on the
thread that calls `eglSwapBuffers`. Decision: the GL backend's `do()` thread owns
ALL GL+EGL, including the window surface and present; the app's window/event
thread only owns native windowing (X11/Win32) and calls Device/Surface methods
that marshal onto the backend thread. `present()` makes the window surface
current on the backend thread, blits, swaps (and may restore surfaceless for
subsequent headless/FBO work). This unifies present onto one GL thread instead of
introducing a second GL-owning thread.

## Bricks

1. **Linux GL windowed present (CI-verifiable).** Implement the seam + the GL
   `backendWindowSurface`; rewire `app/window_linux.go` to create a `gpu.Device`
   (GL) + `CreateWindowSurface(x11 display, window)` and present each frame via
   `PresentImage`; delete `app/ctx_egl_linux.go` and the inline textured-quad GL
   code. Gate with an extended `TestX11WindowedPresent` that drives SEVERAL frames
   + a RESIZE (a one-frame test won't surface a per-thread current-context bug).
   `gpu/gl`+`gpu/ctx/egl` stay (windows still uses them).
2. **Windows GL windowed present (build-only verified — user-approved).** Make
   `gpu/backend_gl.go` cross-platform (linux + windows): load
   `libEGL.so.1`/`libGLESv2.so.2` on linux, ANGLE `libEGL.dll`/`libGLESv2.dll` on
   windows, by GOOS. Wire `app/window_windows.go` to the Device API; delete
   `app/ctx_gl_windows.go` and its textured-quad. Windows present cannot be
   runtime-verified here; it ships build-only-verified and the known-good ANGLE
   path is removed (user chose "migrate both now, delete everything", 2026-06-29).
3. **Delete the old stack.** Remove `gpu/gl` and `gpu/ctx/egl` (now unused),
   including `egl_windows.go`'s `gl.LibGLESv2` dependency. Confirm
   `GOOS={linux,windows,darwin} CGO_ENABLED=0 go build ./...` and the Xvfb job
   stay green.

## Verification

- Linux: extended `TestX11WindowedPresent` (multi-frame + resize) under the
  Xvfb + Mesa `gl-probe` job (already `pull_request`-gated). This is the proof of
  both the present path and the thread/context model.
- Windows: `GOOS=windows CGO_ENABLED=0 go build ./...` only; runtime is a real-
  hardware follow-up.
- Darwin: unchanged; native build + existing tests.

## Out of scope

- Migrating darwin present onto the seam (keep the Metal blit; future target).
- Rendering GPU frames directly into the drawable without the CPU `*image.RGBA`
  readback (the seam allows it via `acquire()`, but the app still presents a CPU
  image for now).
- The pre-existing resize-freeze FIXME in the linux draw loop.
