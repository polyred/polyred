// Copyright 2026 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package render

import (
	"image"
	"unsafe"

	"poly.red/gpu"
	"poly.red/gpu/shader/gpumath/kernels"
)

// gpuGammaCorrect applies sRGB gamma correction to the R, G, B channels of img
// on the GPU, in place, leaving alpha unchanged. This is the renderer's gamma
// pass (shader.GammaCorrection) offloaded to the poly.red/gpu abstraction, using
// the author-once kernels.SRGB (kernels.SRGBSrc). It matches the CPU LUT path
// within +/-1 on 8-bit output.
func gpuGammaCorrect(dev *gpu.Device, img *image.RGBA) error {
	mod, err := kernelModule(dev, kernels.SRGBSrc, "SRGB")
	if err != nil {
		return err
	}
	layout := dev.NewBindGroupLayout(
		gpu.BindGroupLayoutEntry{Binding: 0, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
		gpu.BindGroupLayoutEntry{Binding: 1, Visibility: gpu.StageCompute, Kind: gpu.StorageBuffer},
	)
	pipe, err := dev.NewComputePipeline(gpu.ComputePipelineDescriptor{
		Layout: dev.NewPipelineLayout(layout), Module: mod, Entry: "SRGB",
	})
	if err != nil {
		return err
	}

	pix := img.Pix
	n := len(pix) / 4
	if n == 0 {
		return nil
	}

	// Pack the R,G,B channels (alpha excluded) into a normalized float buffer.
	in := make([]float32, n*3)
	for i := 0; i < n; i++ {
		in[i*3+0] = float32(pix[i*4+0]) / 255
		in[i*3+1] = float32(pix[i*4+1]) / 255
		in[i*3+2] = float32(pix[i*4+2]) / 255
	}
	count := n * 3

	inBuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: count * 4, Usage: gpu.BufferStorage | gpu.BufferCopyDst, Data: f32bytes(in)})
	if err != nil {
		return err
	}
	defer inBuf.Release()
	outBuf, err := dev.NewBuffer(gpu.BufferDescriptor{Size: count * 4, Usage: gpu.BufferStorage | gpu.BufferMapRead})
	if err != nil {
		return err
	}
	defer outBuf.Release()

	bg := dev.NewBindGroup(layout,
		gpu.BindGroupEntry{Binding: 0, Buffer: inBuf},
		gpu.BindGroupEntry{Binding: 1, Buffer: outBuf})

	enc := dev.NewCommandEncoder()
	cp := enc.BeginComputePass()
	cp.SetPipeline(pipe)
	cp.SetBindGroup(0, bg)
	cp.Dispatch(count, 1, 1)
	cp.End()
	dev.Queue().Submit(enc.Finish())
	dev.Queue().WaitIdle()

	out := unsafe.Slice((*float32)(unsafe.Pointer(&outBuf.Bytes()[0])), count)
	for i := 0; i < n; i++ {
		pix[i*4+0] = toU8(out[i*3+0])
		pix[i*4+1] = toU8(out[i*3+1])
		pix[i*4+2] = toU8(out[i*3+2])
		// alpha (pix[i*4+3]) left unchanged
	}
	return nil
}

func toU8(v float32) uint8 {
	x := v*255 + 0.5
	if x < 0 {
		return 0
	}
	if x > 255 {
		return 255
	}
	return uint8(x)
}

func f32bytes(d []float32) []byte {
	return unsafe.Slice((*byte)(unsafe.Pointer(&d[0])), len(d)*4)
}
