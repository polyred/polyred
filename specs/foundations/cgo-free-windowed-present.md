---
title: "cgo-free windowed present: archive the cgo windowing toy"
status: in progress (darwin done; linux next)
depends_on:
  - foundations/gpu-windowed-present.md
affects:
  - gpu/ctx/ca
  - app
  - gpu/gl
  - gpu/ctx/egl
created: 2026-06-21
updated: 2026-06-29
author: changkun
dispatched_task_id: null
---

# cgo-free windowed present: archive the cgo windowing toy

## Overview

The last old design still standing: `app/` windowed present runs on the cgo
windowing toy (`app/window_darwin.go` inline Obj-C, `gpu/ctx/ca` cgo CAMetalLayer,
`gpu/ctx/egl` + `gpu/gl` cgo EGL/GL), even though compute + headless render are
already cgo-free on the Device API (purego). "cgo-free is a hard requirement" and
"replace the toy" are therefore half-done. This arc ports windowed present to
cgo-free purego/objc, brick by brick, darwin-first, then archives the cgo toy.

Working model (decided 2026-06-21): the maintainer verifies on-screen (a window
appearing + presenting needs a display, unverifiable in CI / by the agent). Each
brick gets the partial offscreen FFI verification that IS possible.

## Bricks

1. **CAMetalLayer ops cgo-free (DONE).** `gpu/ctx/ca/metal_layer.go` rewritten from
   cgo (deleted `metal_layer.m`) to purego/objc (same approach as `gpu/mtl`):
   `setDevice:`/`setPixelFormat:`/`setMaximumDrawableCount:`/`setDisplaySyncEnabled:`/
   `setDrawableSize:` (CGSize by value)/`nextDrawable`/drawable `texture`. The layer
   pointer still comes from the native view (the platform window, still cgo). The
   command-buffer `presentDrawable:` was already purego in `gpu/mtl`. Result:
   `gpu/ctx/ca` is cgo-free. Verified: `metal_layer_darwin_test.go` exercises the
   ops on an off-screen CAMetalLayer (setters do not crash, pixel format
   round-trips); on-screen present is a maintainer check (run polyred/polywine).
2. **NSWindow / NSView / event loop cgo-free (darwin) — DONE.** Ported
   `app/window_darwin.go` from inline Obj-C to purego/objc (NSApplication, NSWindow,
   NSView, the run loop, delegates, `layerForView`), then wired mouse/keyboard
   events, window focus, modifier-key reporting (`flagsChanged:`), trackpad scroll
   scaling, and Shift+left-drag pan controls. Darwin app builds `CGO_ENABLED=0` and
   imports neither `gpu/gl` nor `gpu/ctx/egl`. On-screen verified by the maintainer.
3. **Linux window cgo-free (X11/EGL/GLES) — NEXT.** The Linux path is the only
   remaining cgo windowing. Three layers, ported bottom-up, each its own commit
   (each package builds independently `GOOS=linux CGO_ENABLED=0`):
   - **(a) `gpu/gl/gl_unix.go` -> purego GLES.** This is the last cgo in package
     `gl` (`gl_windows.go` is already syscall-based). Reuse the purego dlopen/sym +
     GLES binding pattern from `gpu/backend_gl.go`. Only the Linux window imports it
     now (darwin dropped it in brick 2), so it can narrow to `linux`.
   - **(b) `gpu/ctx/egl` -> purego EGL incl. `eglCreateWindowSurface`.** The compute
     backend's EGL in `gpu/backend_gl.go` is surfaceless; the window needs a real
     window surface bound to the X11 window + `eglSwapBuffers` present.
   - **(c) `app/window_linux.go` -> purego X11 (Xlib).** dlopen `libX11.so.6`:
     `XOpenDisplay`/`XCreateWindow`/`XMapWindow`/`XNextEvent`/atoms/event structs.
     Mirror the existing cgo structure 1:1; do NOT fix the pre-existing resize-freeze
     FIXME (`window_linux.go:187`) in the same diff (one concern at a time).
   Windows is ALREADY cgo-free (`GOOS=windows CGO_ENABLED=0 go build ./app/...` is
   clean; it uses syscall-to-DLL). No Windows port work is needed; on-Windows runtime
   testing is a separate, currently-unreachable concern.
4. **Archive `gpu/ctx/ca` cgo remnants + the cgo bits of `gpu/gl`/`gpu/ctx/egl`**
   once Linux is cgo-free (darwin + windows already are).

## Verification (Linux)

Unlike darwin (on-screen only), the Linux window is CI-gatable. `polyred.yml`
already runs `Xvfb :99` + Mesa GL (`xvfb`, `xorg-dev`, `libgl1-mesa-dev`), and
`gl-probe.yml` proves cgo-free purego EGL/GLES on llvmpipe. The Linux port lands a
windowed-present smoke test (open an X11 window -> create EGL window surface ->
draw one frame -> assert no crash / read back pixels) that runs under Xvfb. The
discriminating unknown the test resolves: whether `eglCreateWindowSurface` on the
X11 window matches a Mesa EGL config under Xvfb (the classic visual/config-matching
snag). Dev loop on darwin: cross-compile `GOOS=linux CGO_ENABLED=0` locally, then
push and let the Xvfb CI job exercise it at runtime.

## Testing Strategy

- Partial offscreen FFI tests per brick (bindings load, basic calls do not crash),
  as in brick 1.
- `CGO_ENABLED=0 go build` of the ported packages confirms they are cgo-free.
- On-screen correctness: the maintainer runs a windowed app (`go run ./cmd/polyred`
  / `./cmd/polywine`) and confirms a window appears and renders.

## Out of scope

- The GPU forward rasterizer arc (separate).
- Linux/Windows window ports until the darwin path is proven (later bricks).

## Deliverable (arc)

Windowed present runs cgo-free on the Device API; the cgo windowing toy
(`gpu/gl`, `gpu/ctx/egl`, the cgo bits of `gpu/ctx/ca` and `app/window_*`) is
archived. The whole GPU + windowing stack is then cgo-free.
