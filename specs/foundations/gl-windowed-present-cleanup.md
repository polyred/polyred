---
title: "GL windowed present via the Device API; retire gpu/gl + gpu/ctx/egl"
status: complete (bricks 1-3 done; linux CI-proven, windows build-only)
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

1. **Linux GL windowed present (CI-verifiable) — DONE, CI-PROVEN.** Implemented the
   seam + the GL `backendWindowSurface`; rewired `app/window_linux.go` onto
   `gpu.Device`(GL) + `CreateWindowSurface` + `PresentImage`; deleted
   `app/ctx_egl_linux.go` + the inline textured-quad. `TestX11WindowedPresent`
   drives several frames + a resize and asserts the presented pixels (red) read
   back, under the Xvfb+Mesa gl-probe job (`--- PASS` on main). Two non-obvious
   runtime fixes CI caught (build-green is not enough here): (a) request a
   window-only EGL config (WINDOW|PBUFFER together has no llvmpipe config -> fell
   back to pbuffer-only, no WINDOW_BIT); (b) the window must use the EGL config's
   EGL_NATIVE_VISUAL_ID (else EGL_BAD_MATCH) AND the device must bind the X11
   Display* via `gpu.WithNativeDisplay` so EGL uses the X11 platform (else
   eglGetDisplay(DEFAULT) picks surfaceless/GBM -> EGL_BAD_NATIVE_WINDOW). The app
   opens the GL device first, reads `Device.WindowVisualID()`, and creates the
   window with that visual (`createX11Window`: XGetVisualInfo + XCreateColormap).
   `gpu/gl`+`gpu/ctx/egl` stay (windows still uses them).
2. **Windows GL windowed present (build-only — user-approved) — DONE.**
   `gpu/backend_gl.go` is now `//go:build linux || windows`. The EGL/GLES lib names
   and loader moved behind build-tagged `glDlopen`/`glDlsym`: linux keeps
   `purego.Dlopen(libEGL.so.1/libGLESv2.so.2, RTLD_NOW|GLOBAL)` byte-identically;
   windows uses `syscall.LoadLibrary`/`GetProcAddress` on ANGLE
   `libEGL.dll`/`libGLESv2.dll` (purego's dlopen is Unix-only; the GL call sites
   stay on `purego.SyscallN`, which IS on windows/amd64). The Vulkan dispatch is
   `openVKBackend` (linux) + a non-linux ErrUnsupported stub; `backend_other.go`
   now excludes windows. `app/window_windows.go` opens `gpu.Device`(GL,
   WithNativeDisplay(hdc)) + `CreateWindowSurface(hwnd)` (no visual matching: ANGLE
   takes the HWND directly) + `PresentImage`; deleted `app/ctx_gl_windows.go`.
   Build-only verified (no Windows CI runtime). Sharpest runtime delta: relies on
   ANGLE to load `d3dcompiler_47.dll` from the search path (the old path
   LoadLibrary'd it explicitly). ANGLE may also not support surfaceless
   `eglMakeCurrent` at init -- a real-hardware concern.
3. **Delete the old stack — DONE.** Removed `gpu/gl` and `gpu/ctx/egl` entirely
   (nothing imports them after bricks 1-2). The GPU backend's own purego EGL/GLES
   in `backend_gl.go` is the single GL stack now; `gpu/ctx/ca`+`gpu/mtl` (darwin
   Metal) untouched. `GOOS={linux,windows,darwin} CGO_ENABLED=0 go build ./...` +
   gpu/app tests pass.

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
