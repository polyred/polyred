// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin

package app

import (
	"image"
	"runtime"
	"testing"

	"github.com/ebitengine/purego/objc"
	"poly.red/gpu/mtl"
)

// TestBlitPresentNoUseAfterFree reproduces the windowed-present SIGSEGV that
// `cmd/polyred show` hit: flush() blits a frame into the drawable, registers a
// command-buffer completion handler, and returns -- draining its autorelease pool --
// before the handler fires on the GPU thread. The old handler then sent -release to
// the (autoreleased, already-freed) command buffer and blit encoder, a use-after-free
// that crashed in objc_msgSend. This drives the same blitPresent path many times,
// each with a per-frame autorelease pool that drains before completion; with the bug
// it SIGSEGVs (aborting the test binary), with the fix every completion handler fires
// cleanly. It blits into an offscreen texture, so it needs no window / CAMetalLayer.
func TestBlitPresentNoUseAfterFree(t *testing.T) {
	// flush() presents on a thread-locked draw goroutine; an autorelease pool must be
	// alloc'd and drained on the same OS thread, so lock this goroutine too (objc
	// calls are scheduling points that would otherwise let it migrate mid-frame).
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	dev, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		t.Skipf("no Metal device: %v", err)
	}
	ctx := &mtlContext{device: dev, queue: dev.MakeCommandQueue()}
	defer ctx.queue.Release()

	const dx, dy = 32, 32
	img := image.NewRGBA(image.Rect(0, 0, dx, dy))
	for i := range img.Pix {
		img.Pix[i] = byte(i)
	}
	dst := dev.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatBGRA8UNorm,
		Width:       dx,
		Height:      dy,
		StorageMode: mtl.StorageModeManaged,
	})
	defer dst.Release()

	// Many frames: each completion handler must fire without crashing. The frame's
	// autorelease pool drains at the end of the closure, before the handler runs --
	// exactly the ordering that made the old release-in-handler a use-after-free.
	const frames = 300
	for i := 0; i < frames; i++ {
		done := make(chan struct{})
		func() {
			pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(selAlloc).Send(selInit)
			defer pool.Send(selRelease)
			ctx.blitPresent(img, dst, nil, func() { close(done) })
		}()
		<-done
	}
}
