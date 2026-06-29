---
title: "cgo-free windowed present: archive the cgo windowing toy"
status: in progress (all platforms cgo-free; linux runtime pending CI)
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
3. **Linux window cgo-free (X11/EGL/GLES) — DONE (runtime pending CI).** Ported
   bottom-up, each layer its own commit (each builds `GOOS=linux CGO_ENABLED=0`):
   - **(a) `gpu/gl/gl_unix.go` -> purego GLES (DONE).** The last cgo in package
     `gl`. Each GL entry point is a typed Go func resolved with
     `purego.RegisterFunc` so the System V AMD64 float ABI is correct
     (`glClearColor`/`glUniform*f` would be corrupted by integer-register
     `SyscallN`). Required/optional symbol split + EXT fallbacks preserved 1:1.
   - **(b) `gpu/ctx/egl` -> purego EGL (DONE).** `egl_linux.go` resolves
     `libEGL.so.1` and calls via `purego.SyscallN` (EGL has no float args). Handle
     types became uintptr, matching the cgo-free `egl_windows.go` sibling, so the
     cross-platform `egl.go` is unchanged. Deleted `egl_x11.go` (dead cgo;
     `NewDisplay` had no callers, its import-time libX11 dlopen was a fragile side
     effect).
   - **(c) `app/window_linux.go` -> purego X11 (DONE).** `libX11.so.6` + 14 Xlib
     entry points via `purego.SyscallN`; `XEvent`/`XSetWindowAttributes`/
     `XTextProperty` mirrored as Go structs with explicit LP64 padding (offsets
     verified field-by-field against `X11/Xlib.h`/`Xutil.h`). Left the pre-existing
     resize-freeze FIXME untouched.
   Bonus bug fix (commit `18aae9f`): commit 3fd9c9d had left the linux build broken
   (duplicate Mod consts in `event_linux.go` vs the untagged `event_mods.go`; main's
   `polyred` CI was red). Deleted `event_linux.go`, added a unit-tested pure
   `x11ModsToLogical` that maps raw X11 state to the logical ModifierKey bits
   (without it, `Contain(ModShift)` never matched on linux and Shift+drag pan was
   dead). The three event sites now use it.

   Windows was ALREADY cgo-free (`GOOS=windows CGO_ENABLED=0 go build ./...` clean;
   syscall-to-DLL). No Windows port work needed; on-Windows runtime testing is a
   separate, currently-unreachable concern.
4. **Archive note.** With (a)-(c) done, the whole repo is cgo-free: `import "C"`
   appears nowhere, and `./...` cross-builds `CGO_ENABLED=0` for linux, windows, and
   darwin. The old cgo toy (`gpu/gl`, `gpu/ctx/egl`, `gpu/ctx/ca`) was *ported* to
   purego rather than deleted, because linux/windows windowed present still ride
   `gpu/gl`+`gpu/ctx/egl` (the textured-quad present), unlike darwin which presents
   via the Device API (`gpu/ctx/ca`+`gpu/mtl`). Migrating linux/windows present onto
   the Device API too (so `gpu/gl`/`egl` could finally be deleted) is a separate,
   larger arc; the cgo-free HARD REQUIREMENT is now met on all platforms.

## Verification

Darwin: on-screen, maintainer-verified (done). Windows: builds cgo-free; runtime
unreachable here. Linux: CI-gatable, and now gated. `TestX11WindowedPresent`
(`app/window_linux_test.go`) drives the real path without the blocking event loop
(open X11 window -> `eglCreateWindowSurface` -> make current -> clear to opaque red
-> `ReadPixels` and assert ~255,0,0,255). Red is sRGB-invariant and channel-specific
(catches channel-swap); `ClearColor` exercises the float ABI; `GetString`/`GetInteger`
exercise out-parameter marshaling. It runs in a new `x11-windowed-present` job in
`gl-probe.yml` (Xvfb + Mesa llvmpipe + libEGL/libGLESv2), now `pull_request`-gated.
It skips cleanly without a display/EGL runtime, so it is a no-op on a bare dev box.
The remaining unknown only a CI run resolves: whether `eglCreateWindowSurface` on the
X11 window matches a Mesa EGL config under Xvfb (the classic visual/config snag).
Dev loop on darwin was: cross-compile `GOOS=linux CGO_ENABLED=0` locally; push for
the Xvfb job to exercise it at runtime.

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
