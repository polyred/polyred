// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin
// +build darwin

package app

import (
	"errors"

	"poly.red/app/internal/mtl"
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

func (ctx *mtlContext) Release() {
	ctx.texture.Release()
	ctx.queue.Release()
}
