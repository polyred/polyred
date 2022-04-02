// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// This is a minimum example that demonstrates GPU compute shader using Metal.
// Inspired from https://adrianhesketh.com/2022/03/31/use-m1-gpu-with-go/

//go:build darwin

package main

import (
	_ "embed"
	"errors"
	"log"
	"unsafe"

	"poly.red/internal/bytes"
	"poly.red/internal/driver/mtl"
	"poly.red/math"
)

//go:embed compute.metal
var compute string

func main() {
	in := math.NewMat[float32](2, 10,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
		0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	)
	out := math.NewMat[float32](1, 10)

	Compute(in, out)

	log.Println(in, out)
}

type params struct {
	WIn, HIn   int32
	WOut, HOut int32
}

func Compute[TIn, TOut math.Type](in math.Mat[TIn], out math.Mat[TOut]) (err error) {
	defer handle(func(er error) {
		err = er
	})

	sizeIn := math.TypeSize[TIn]()
	sizeOut := math.TypeSize[TOut]()

	device := try(mtl.CreateSystemDefaultDevice())
	fn := try(try(device.MakeLibrary(compute, mtl.CompileOptions{
		LanguageVersion: mtl.LanguageVersion1_1,
	})).MakeFunction("main0"))
	cps := try(device.MakeComputePipelineState(fn))

	bufIn := device.MakeBuffer(unsafe.Pointer(&in.Data[0]), uintptr(sizeIn*len(in.Data)), mtl.ResourceStorageModeShared)
	bufOut := device.MakeBuffer(nil, uintptr(sizeOut*len(in.Data)), mtl.ResourceStorageModeShared)

	p := &params{
		WIn:  int32(in.Row),
		HIn:  int32(in.Col),
		WOut: int32(out.Row),
		HOut: int32(out.Col),
	}

	cmdBuffer := device.MakeCommandQueue().MakeCommandBuffer()
	computeEncoder := cmdBuffer.MakeComputeCommandEncoder()
	computeEncoder.SetComputePipelineState(cps)
	computeEncoder.SetBytes(bytes.FromStruct(p), 0)
	computeEncoder.SetBuffer(bufIn, 0, 1)
	computeEncoder.SetBuffer(bufOut, 0, 2)

	threadsPerGrid := mtl.Size{
		Width:  in.Row,
		Height: in.Col,
		Depth:  1,
	}
	w := cps.ThreadExecutionWidth()
	h := cps.MaxTotalThreadsPerThreadgroup() / w
	threadsPerThreadgroup := mtl.Size{
		Width:  w,
		Height: h,
		Depth:  1,
	}
	computeEncoder.DispatchThreads(threadsPerGrid, threadsPerThreadgroup)
	computeEncoder.EndEncoding()
	cmdBuffer.Commit()
	cmdBuffer.WaitUntilCompleted()
	copy(out.Data, unsafe.Slice((*TOut)(bufOut.Content()), len(out.Data)))
	return nil
}

func try[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func handle(f func(err error)) {
	if r := recover(); r != nil {
		var err error
		switch x := r.(type) {
		case string:
			err = errors.New(x)
		case error:
			err = x
		default:
			err = errors.New("unknown panic")
		}
		f(err)
	}
	return
}
