// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package app

import (
	"errors"

	"poly.red/internal/driver/mtl"
)

type mtlContext struct {
	layer   caMetalLayer
	device  mtl.Device
	queue   mtl.CommandQueue
	texture mtl.Texture
}

func newMtlContext(cfg *config, layer caMetalLayer) (*mtlContext, error) {
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

func (ctx *mtlContext) Release() {
	ctx.texture.Release()
	ctx.queue.Release()
}
