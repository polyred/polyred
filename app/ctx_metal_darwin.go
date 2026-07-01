// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"errors"
	"image"

	"poly.red/gpu/ctx/ca"
	"poly.red/gpu/mtl"
)

type mtlContext struct {
	layer   ca.MetalLayer
	device  mtl.Device
	queue   mtl.CommandQueue
	texture mtl.Texture
}

func newMtlContext(cfg *config, layer ca.MetalLayer) (*mtlContext, error) {
	device, err := mtl.CreateSystemDefaultDevice()
	if err != nil {
		return nil, errors.New("create MTLCreateSystemDefaultDevice failed")
	}

	layer.SetDevice(device)
	layer.SetPixelFormat(mtl.PixelFormatBGRA8UNorm)
	layer.SetMaximumDrawableCount(3)
	layer.SetDrawableSize(cfg.size.X, cfg.size.Y)
	layer.SetDisplaySyncEnabled(true)

	return &mtlContext{
		device: device,
		queue:  device.MakeCommandQueue(),
		layer:  layer,
		texture: device.MakeTexture(mtl.TextureDescriptor{
			PixelFormat: mtl.PixelFormatBGRA8UNorm,
			Width:       cfg.size.X,
			Height:      cfg.size.Y,
			StorageMode: mtl.StorageModeManaged,
		}),
	}, nil
}

func (m *mtlContext) Resize(w, h int) {
	m.layer.SetDrawableSize(w, h)
	m.texture.Release()
	m.texture = m.device.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatBGRA8UNorm,
		Width:       w,
		Height:      h,
		StorageMode: mtl.StorageModeManaged,
	})
}

// blitPresent stages img into a fresh texture, blits it into dst on the GPU, runs
// present (e.g. cb.PresentDrawable(drawable)) if non-nil, and releases the staging
// texture once the GPU finishes -- onComplete runs first, on the completion (GPU)
// thread.
//
// Object ownership is the crux and was the cause of a use-after-free SIGSEGV: the
// command buffer (`commandBuffer`) and blit encoder (`blitCommandEncoder`) are
// AUTORELEASED objects, owned by the caller's autorelease pool, which drains when the
// caller returns -- BEFORE this completion handler fires on the GPU thread. The old
// code Released cb/bce (and dst, which the drawable owns) in the handler, so it sent
// -release to freed objects and crashed. Only `tex`, from newTextureWithDescriptor:
// (+1, owned), is released here; the autoreleased objects must NOT be.
func (c *mtlContext) blitPresent(img *image.RGBA, dst mtl.Texture, present func(cb mtl.CommandBuffer), onComplete func()) {
	dx, dy := img.Bounds().Dx(), img.Bounds().Dy()
	tex := c.device.MakeTexture(mtl.TextureDescriptor{
		PixelFormat: mtl.PixelFormatBGRA8UNorm,
		Width:       dx,
		Height:      dy,
		StorageMode: mtl.StorageModeManaged,
	})
	tex.ReplaceRegion(mtl.RegionMake2D(0, 0, dx, dy), 0, img.Pix, uintptr(4*dx))

	cb := c.queue.MakeCommandBuffer()
	bce := cb.MakeBlitCommandEncoder()
	bce.CopyFromTexture(tex, 0, 0, mtl.Origin{},
		mtl.Size{Width: dx, Height: dy, Depth: 1},
		dst, 0, 0, mtl.Origin{})
	bce.EndEncoding()
	if present != nil {
		present(cb)
	}
	cb.AddCompletedHandler(func() {
		if onComplete != nil {
			onComplete()
		}
		tex.Release() // the only owned (+1) object; cb, bce and dst are autoreleased
	})
	cb.Commit()
}

func (ctx *mtlContext) Release() {
	ctx.texture.Release()
	ctx.queue.Release()
}
