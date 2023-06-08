// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

package tests

import (
	_ "embed"

	"log"
	"unsafe"

	"poly.red/gpu/mtl"
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
	mathLib = try(newMathLibrary())
}

func add[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	if m1.Row != m2.Row || m1.Col != m2.Col {
		panic("gpu: input matrix for Add have different dimensions")
	}

	a := device.MakeBuffer(unsafe.Pointer(&m1.Data[0]), uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer a.Release()
	b := device.MakeBuffer(unsafe.Pointer(&m2.Data[0]), uintptr(math.TypeSize[T]()*len(m2.Data)), mtl.ResourceStorageModeShared)
	defer b.Release()
	out := device.MakeBuffer(nil, uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer out.Release()

	cq := device.MakeCommandQueue()
	defer cq.Release()
	cb := cq.MakeCommandBuffer()
	defer cb.Release()

	ce := cb.MakeComputeCommandEncoder()
	ce.SetComputePipelineState(mathLib.funcAdd.cps)
	ce.SetBuffer(a, 0, 0)
	ce.SetBuffer(b, 0, 1)
	ce.SetBuffer(out, 0, 2)

	threadGroupSize := mathLib.funcAdd.cps.MaxTotalThreadsPerThreadgroup()
	if threadGroupSize > len(m1.Data) {
		threadGroupSize = len(m1.Data)
	}
	ce.DispatchThreads(mtl.Size{Width: len(m1.Data), Height: 1, Depth: 1},
		mtl.Size{Width: threadGroupSize, Height: 1, Depth: 1})

	ce.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	data := make([]T, len(m1.Data))
	copy(data, unsafe.Slice((*T)(out.Content()), len(m1.Data)))
	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: data,
	}
}

func sub[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	if m1.Row != m2.Row || m1.Col != m2.Col {
		panic("gpu: input matrix for Add have different dimensions")
	}

	a := device.MakeBuffer(unsafe.Pointer(&m1.Data[0]), uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer a.Release()
	b := device.MakeBuffer(unsafe.Pointer(&m2.Data[0]), uintptr(math.TypeSize[T]()*len(m2.Data)), mtl.ResourceStorageModeShared)
	defer b.Release()
	out := device.MakeBuffer(nil, uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer out.Release()

	cq := device.MakeCommandQueue()
	defer cq.Release()
	cb := cq.MakeCommandBuffer()
	defer cb.Release()

	ce := cb.MakeComputeCommandEncoder()
	ce.SetComputePipelineState(mathLib.funcSub.cps)
	ce.SetBuffer(a, 0, 0)
	ce.SetBuffer(b, 0, 1)
	ce.SetBuffer(out, 0, 2)

	threadGroupSize := mathLib.funcSub.cps.MaxTotalThreadsPerThreadgroup()
	if threadGroupSize > len(m1.Data) {
		threadGroupSize = len(m1.Data)
	}
	ce.DispatchThreads(mtl.Size{Width: len(m1.Data), Height: 1, Depth: 1},
		mtl.Size{Width: threadGroupSize, Height: 1, Depth: 1})

	ce.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	data := make([]T, len(m1.Data))
	copy(data, unsafe.Slice((*T)(out.Content()), len(m1.Data)))
	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: data,
	}
}

func sqrt[T DataType](m math.Mat[T]) math.Mat[T] {
	a := device.MakeBuffer(unsafe.Pointer(&m.Data[0]), uintptr(math.TypeSize[T]()*len(m.Data)), mtl.ResourceStorageModeShared)
	defer a.Release()
	out := device.MakeBuffer(nil, uintptr(math.TypeSize[T]()*len(m.Data)), mtl.ResourceStorageModeShared)
	defer out.Release()

	cq := device.MakeCommandQueue()
	defer cq.Release()
	cb := cq.MakeCommandBuffer()
	defer cb.Release()

	ce := cb.MakeComputeCommandEncoder()
	ce.SetComputePipelineState(mathLib.funcSqrt.cps)
	ce.SetBuffer(a, 0, 0)
	ce.SetBuffer(out, 0, 1)

	threadGroupSize := mathLib.funcSqrt.cps.MaxTotalThreadsPerThreadgroup()
	if threadGroupSize > len(m.Data) {
		threadGroupSize = len(m.Data)
	}
	ce.DispatchThreads(mtl.Size{Width: len(m.Data), Height: 1, Depth: 1},
		mtl.Size{Width: threadGroupSize, Height: 1, Depth: 1})

	ce.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	data := make([]T, len(m.Data))
	copy(data, unsafe.Slice((*T)(out.Content()), len(m.Data)))
	return math.Mat[T]{
		Row:  m.Row,
		Col:  m.Col,
		Data: data,
	}
}

type params struct {
	WidthA  int
	HeightA int
	WidthB  int
}

func mul[T DataType](m1, m2 math.Mat[T]) math.Mat[T] {
	if m1.Col != m2.Row {
		panic("math: mismatched matrix dimension")
	}

	a := device.MakeBuffer(unsafe.Pointer(&m1.Data[0]), uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer a.Release()
	b := device.MakeBuffer(unsafe.Pointer(&m2.Data[0]), uintptr(math.TypeSize[T]()*len(m2.Data)), mtl.ResourceStorageModeShared)
	defer b.Release()
	out := device.MakeBuffer(nil, uintptr(math.TypeSize[T]()*len(m1.Data)), mtl.ResourceStorageModeShared)
	defer out.Release()

	cq := device.MakeCommandQueue()
	defer cq.Release()
	cb := cq.MakeCommandBuffer()
	defer cb.Release()

	ce := cb.MakeComputeCommandEncoder()
	ce.SetComputePipelineState(mathLib.funcMul.cps)
	ce.SetBuffer(a, 0, 0)
	ce.SetBuffer(b, 0, 1)
	ce.SetBuffer(out, 0, 2)

	pp := params{
		WidthA:  m1.Col,
		HeightA: m1.Row,
		WidthB:  m2.Col,
	}
	dp := device.MakeBuffer(unsafe.Pointer(&pp), uintptr(unsafe.Sizeof(params{})), mtl.ResourceStorageModeShared)
	ce.SetBuffer(dp, 0, 3)
	ce.DispatchThreads(
		mtl.Size{Width: int(pp.WidthB * pp.HeightA), Height: 1, Depth: 1},
		mtl.Size{Width: 1, Height: 1, Depth: 1})

	ce.EndEncoding()
	cb.Commit()
	cb.WaitUntilCompleted()

	data := make([]T, len(m1.Data))
	copy(data, unsafe.Slice((*T)(out.Content()), pp.WidthB*pp.HeightA))
	return math.Mat[T]{
		Row:  m1.Row,
		Col:  m1.Col,
		Data: data,
	}
}

type computeFunc struct {
	fn  mtl.Function
	cps mtl.ComputePipelineState
}

type shaderFn struct {
	lib      mtl.Library
	funcAdd  computeFunc
	funcSub  computeFunc
	funcSqrt computeFunc
	funcMul  computeFunc
}

var (
	//go:embed shaders/math.metal
	mathMetal string
	mathLib   Function
)

func newMathLibrary() (fn Function, err error) {
	defer handle(func(er error) { err = er })
	lib := try(device.MakeLibrary(mathMetal, mtl.CompileOptions{
		LanguageVersion: mtl.LanguageVersion2_4,
	}))

	fn = Function{shaderFn{
		lib: lib,
	}}

	funcAdd := computeFunc{}
	funcAdd.fn = try(lib.MakeFunction("add0"))
	funcAdd.cps = try(device.MakeComputePipelineState(funcAdd.fn))
	fn.funcAdd = funcAdd

	funcSub := computeFunc{}
	funcSub.fn = try(lib.MakeFunction("sub0"))
	funcSub.cps = try(device.MakeComputePipelineState(funcSub.fn))
	fn.funcSub = funcSub

	funcSqrt := computeFunc{}
	funcSqrt.fn = try(lib.MakeFunction("sqrt0"))
	funcSqrt.cps = try(device.MakeComputePipelineState(funcSqrt.fn))
	fn.funcSqrt = funcSqrt

	funcMul := computeFunc{}
	funcMul.fn = try(lib.MakeFunction("mul0"))
	funcMul.cps = try(device.MakeComputePipelineState(funcMul.fn))
	fn.funcMul = funcMul
	return
}
