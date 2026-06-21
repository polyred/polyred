---
title: "cgo-free windowed present: archive the cgo windowing toy"
status: in progress (brick 1 done)
depends_on:
  - foundations/gpu-windowed-present.md
affects:
  - gpu/ctx/ca
  - app
  - gpu/gl
  - gpu/ctx/egl
created: 2026-06-21
updated: 2026-06-21
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
2. **NSWindow / NSView / event loop cgo-free (darwin).** Port `app/window_darwin.go`
   from inline Obj-C to purego/objc (NSApplication, NSWindow, NSView, the run loop,
   delegates, `layerForView`). The hardest brick; large FFI; on-screen only.
3. **Archive `gpu/ctx/ca` cgo remnants + `gpu/gl` + `gpu/ctx/egl`** once their cgo
   users (app windows) are cgo-free. Linux/Windows windows (X11/EGL/Win32) port
   analogously (separate bricks, those platforms).

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
