// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package gpu

import (
	_ "embed"

	"log"
	"unsafe"

	"poly.red/internal/driver/mtl"
	"poly.red/math"
)

var device mtl.Device

func init() {
	defer handle(func(err error) {
		if err != nil {
			log.Println(err)
		}
	})

	device = try(mtl.CreateSystemDefaultDevice())
	addFn = try(newAddShader())
}

func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	if m1.Row != m2.Row || m1.Col != m2.Col {
		panic("gpu: input matrix for Add have different dimensions")
	}

	a := device.MakeBuffer(unsafe.Pointer(&m1.Data[0]), uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	b := device.MakeBuffer(unsafe.Pointer(&m2.Data[0]), uintptr(math.TypeSize[T]()*len(m2.Data)), mtl.ResourceStorageModeShared)
	out := device.MakeBuffer(nil, uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)

	cb := device.MakeCommandQueue().MakeCommandBuffer()
	ce := cb.MakeComputeCommandEncoder()
	ce.SetComputePipelineState(addFn.cps)
	ce.SetBuffer(a, 0, 0)
	ce.SetBuffer(b, 0, 1)
	ce.SetBuffer(out, 0, 2)

	threadGroupSize := addFn.cps.MaxTotalThreadsPerThreadgroup()
	if threadGroupSize > len(m1.Data) {
		threadGroupSize = len(m1.Data)
	}
	ce.DispatchThreads(mtl.Size{Width: len(m1.Data), Height: 1, Depth: 1},
		mtl.Size{Width: threadGroupSize, Height: 1, Depth: 1})

	ce.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: unsafe.Slice((*T)(out.Content()), len(m1.Data)),
	}
}

type shaderFn struct {
	fn  mtl.Function
	cps mtl.ComputePipelineState
}

var (
	//go:embed add.metal
	addMetal string
	addFn    Function
)

func newAddShader() (sf Function, err error) {
	defer handle(func(er error) { err = er })
	fn := try(try(device.MakeLibrary(addMetal, mtl.CompileOptions{
		LanguageVersion: mtl.LanguageVersion2_4,
	})).MakeFunction("main0"))
	cps := try(device.MakeComputePipelineState(fn))
	sf = Function{shaderFn{fn, cps}}
	return
}
