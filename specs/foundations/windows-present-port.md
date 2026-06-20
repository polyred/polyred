---
title: Windows present path: port to the modern textured-quad GLES blit
status: implemented (build); runtime deferred
depends_on:
  - foundations/gpu-phase1-foundation.md
affects:
  - app/ctx_gl_windows.go
  - app/window_windows.go
  - gpu/gl/gl_windows.go
effort: small
created: 2026-06-21
updated: 2026-06-21
author: changkun
dispatched_task_id: null
---

# Windows present path: port to the modern textured-quad GLES blit

## Overview

The Windows `app` window does not build. `app/ctx_gl_windows.go` and
`app/window_windows.go` call a package-level, immediate-mode GL API
(`gl.MakeCurrent`, `gl.DrawBuffer`, `gl.FRONT`, `gl.PixelZoom`,
`gl.RasterPos2d`, `gl.Viewport`, `gl.DrawPixels`, `gl.Finish`) that the
restructured GLES `gpu/gl` package no longer provides. `gpu/gl` now exposes GLES
2/3 entry points as methods on `*gl.Functions`. This is the only red in CI
(macOS + Linux are green). The Linux path (`app/window_linux.go`) was already
modernized to present a frame by uploading the rendered `*image.RGBA` to a
texture and drawing a full-screen textured quad via an EGL (ANGLE) context. This
spec ports the Windows path to the same approach, reusing the **already-present**
Windows EGL backend (`gpu/ctx/egl/egl_windows.go`, ANGLE via `libEGL.dll` +
`libGLESv2.dll`).

## Current State

- `app/window_linux.go` (reference): `run()` opens an X11 window, builds a
  `*x11Context` (`app/ctx_egl_linux.go`) wrapping `*egl.Context` + `*gl.Functions`,
  then `go w.draw(app)` and signals `w.ready`. `draw()` locks the OS thread,
  `ctx.Lock()` (eglMakeCurrent), sets up a VBO of a `-1..+1` quad with UVs, a
  GLES program (`vert`/`frag` constants), attribute pointers, and one texture;
  each frame `flush()` uploads `img.Pix` with `TexImage2D` and draws a
  `TRIANGLE_STRIP`, then `ctx.Present()` (eglSwapBuffers).
- `gpu/ctx/egl/egl.go` (`//go:build linux || windows`) + `egl_windows.go` already
  provide the identical `*egl.Context` API on Windows (`NewContext`,
  `CreateSurface`, `MakeCurrent`, `ReleaseCurrent`, `Present`, `EnableVSync`,
  `Release`). ANGLE accepts the window's `HDC` as the native display and the
  `HWND` as the native window.
- `gpu/gl/gl_windows.go` already implements every `*Functions` method the Linux
  `draw`/`flush` use (CreateBuffer, BindBuffer, BufferData, UseProgram,
  EnableVertexAttribArray, VertexAttribPointer, CreateTexture, BindTexture,
  TexImage2D, TexParameteri, DrawArrays, Finish, Viewport, …) **except**
  `GetAttribLocation`, which the Linux `draw` calls and which is absent on
  Windows.
- `app/window.go` `window` struct already has `ready`, `resize`, `fontDrawer`,
  the fields the Linux goroutine model needs.

## Architecture

Mirror Linux exactly, substituting the Windows native handles:

| Concern | Linux | Windows (this port) |
| --- | --- | --- |
| Native display | `C.Display*` | window `HDC` (`egl.NativeDisplayType(hdc)`) |
| Native window | `C.Window` | window `HWND` (`egl.NativeWindowType(hwnd)`) |
| Context wrapper | `x11Context` | `winContext` (same methods) |
| Draw driver | `go w.draw(app)` + X11 event loop in `main` | `go w.draw(app)` + Win32 message pump in `event()` |
| Present | textured-quad blit | identical textured-quad blit |

## Components

### `gpu/gl/gl_windows.go`: add `GetAttribLocation`

Add the missing method + its lazy proc, mirroring the existing
`GetUniformLocation` (which already uses `cString` + `issue34474KeepAlive`):

```go
_glGetAttribLocation = LibGLESv2.NewProc("glGetAttribLocation")

func (c *Functions) GetAttribLocation(p Program, name string) Attrib {
	cname := cString(name)
	c0 := &cname[0]
	a, _, _ := syscall.Syscall(_glGetAttribLocation.Addr(), 2, uintptr(p.V), uintptr(unsafe.Pointer(c0)), 0)
	issue34474KeepAlive(c0)
	return Attrib(a)
}
```

### `app/ctx_gl_windows.go`: `winContext` mirroring `x11Context`

Replace the immediate-mode `glContext` with a `winContext` holding `*egl.Context`
+ `*gl.Functions`, with `newWinContext(w)/Release/Refresh/Lock/Unlock/Present`
identical in shape to `ctx_egl_linux.go`. The display comes from the window's
`hdc`; the surface from the `hwnd`.

### `app/window_windows.go`: goroutine draw + textured-quad flush

- `osWindow.ctx` field type changes `glContext` → `*winContext`.
- `run()`: after `GetDC`, build the context (`newWinContext`), `ctx.Refresh()`,
  then `go w.draw(app)` and run the message pump `w.event()` (unchanged). Remove
  the per-`WM_PAINT` `w.draw(app)` call (it would block the pump and re-enter
  draw); the single background `draw` goroutine owns rendering, as on Linux.
- `draw()`: replace the body with the Linux `draw()` body (LockOSThread,
  `ctx.Lock()`, build VBO/program/attribs/texture once, ticker loop, `flush` +
  `ctx.Present()`), keeping the Windows `resize`/fps handling.
- `flush()`: replace `RasterPos2d/Viewport/DrawPixels/Finish` with the Linux
  `flush()` (`Viewport`, `TexImage2D`, four `TexParameteri`, `DrawArrays`,
  `Finish`) through `w.win.ctx.gl`.
- Reuse the same `vert`/`frag` GLES shader source as Linux (define once for the
  package or duplicate in the Windows file, since they are `//go:build`-exclusive, so
  no symbol clash; duplicate to keep each file self-contained).

## Error Handling

Context creation and surface creation already return errors; `run()` panics on
them exactly as Linux/the current Windows code does (window creation is fatal).
No new failure modes.

## Testing Strategy

- **Build is the gate.** CI's Windows job runs `go build ./...` + `go vet
  -unsafeptr=false ./...` + `go test`. The fix is verified by the Windows CI job
  turning green (the build currently fails with 8 `undefined: gl.*` errors). This
  is the authoritative verification loop: this change cannot be compiled from the
  darwin dev machine (`GOOS=windows` cross-build of the `app` cgo/syscall stack),
  so iterate via push → read the Windows job → adjust.
- **Cross-compile smoke (best-effort):** `GOOS=windows GOARCH=amd64
  CGO_ENABLED=0 go build ./app ./gpu/gl ./gpu/ctx/egl` from darwin catches the
  `undefined` symbols and signature mismatches even though it cannot link a real
  binary; run it locally before pushing.
- **Runtime windowing (deferred):** actually displaying a window + pumping
  Win32 messages needs a Windows desktop session and is out of scope here; this
  port makes Windows build and structurally mirrors the verified Linux present
  path. Documented as such in `specs/README.md`.
