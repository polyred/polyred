// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package mtl provides the absolute minimum access to
// Apple's Metal API for drawing images on a CAMetalLayer.
// https://developer.apple.com/documentation/metal
//
// This package requires macOS version 10.13 or newer.
package mtl

/*
// In macOS, in order for the system to provide a default Metal device
// object, you must link to the Core Graphics framework. You usually
// need to do this explicitly if you’re writing apps that don’t use
// graphics by default, such as command line tools.

#cgo CFLAGS: -Werror -fmodules -x objective-c
#cgo LDFLAGS: -framework Metal -framework CoreGraphics
#include "mtl.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// Device is abstract representation of the GPU that
// serves as the primary interface for a Metal app.
// https://developer.apple.com/documentation/metal/mtldevice.
type Device struct {
	device unsafe.Pointer

	// Headless indicates whether a device is configured as headless.
	Headless bool

	// LowPower indicates whether a device is low-power.
	LowPower bool

	// Removable determines whether or not a GPU is removable.
	Removable bool

	// RegistryID is the registry ID value for the device.
	RegistryID uint64

	// Name is the name of the device.
	Name string
}

// CreateSystemDefaultDevice returns the preferred system default Metal device.
// https://developer.apple.com/documentation/metal/1433401-mtlcreatesystemdefaultdevice.
func CreateSystemDefaultDevice() (Device, error) {
	d := C.CreateSystemDefaultDevice()
	if d.Device == nil {
		return Device{}, errors.New("metal is not supported on this system")
	}

	return Device{
		device:     d.Device,
		Headless:   bool(d.Headless),
		LowPower:   bool(d.LowPower),
		Removable:  bool(d.Removable),
		RegistryID: uint64(d.RegistryID),
		Name:       C.GoString(d.Name),
	}, nil
}

// Available returns true if the current macOS supports Metal.
func (d Device) Available() bool {
	return d.Device() != nil
}

// Device returns the underlying id<MTLDevice> pointer.
func (d Device) Device() unsafe.Pointer { return d.device }

// MakeCommandQueue creates a serial command submission queue.
// https://developer.apple.com/documentation/metal/mtldevice/1433388-makecommandqueue.
func (d Device) MakeCommandQueue() CommandQueue {
	return CommandQueue{C.Device_MakeCommandQueue(d.device)}
}

// Region is a rectangular block of pixels in an image or texture,
// defined by its upper-left corner and its size.
// https://developer.apple.com/documentation/metal/mtlregion.
type Region struct {
	Origin Origin // The location of the upper-left corner of the block.
	Size   Size   // The size of the block.
}

// Origin represents the location of a pixel in an image or texture relative
// to the upper-left corner, whose coordinates are (0, 0).
// https://developer.apple.com/documentation/metal/mtlorigin.
type Origin struct{ X, Y, Z int }

// Size represents the set of dimensions that declare the size of an object,
// such as an image, texture, threadgroup, or grid.
// https://developer.apple.com/documentation/metal/mtlsize.
type Size struct{ Width, Height, Depth int }

// RegionMake2D returns a 2D, rectangular region for image or texture data.
// https://developer.apple.com/documentation/metal/1515675-mtlregionmake2d.
func RegionMake2D(x, y, width, height int) Region {
	return Region{
		Origin: Origin{x, y, 0},
		Size:   Size{width, height, 1},
	}
}

// MakeTexture creates a texture object with privately owned storage
// that contains texture state.
// https://developer.apple.com/documentation/metal/mtldevice/1433425-maketexture.
func (d Device) MakeTexture(td TextureDescriptor) Texture {
	descriptor := C.struct_TextureDescriptor{
		PixelFormat: C.uint16_t(td.PixelFormat),
		Width:       C.uint_t(td.Width),
		Height:      C.uint_t(td.Height),
		StorageMode: C.uint8_t(td.StorageMode),
	}
	texture := C.Device_MakeTexture(d.device, descriptor)
	return Texture{
		texture: texture,
		width:   int(C.MTLTexture_GetWidth(texture)),
		height:  int(C.MTLTexture_GetHeight(texture)),
	}
}

// StorageMode defines defines the memory location and access permissions of a resource.
// https://developer.apple.com/documentation/metal/mtlstoragemode.
type StorageMode uint8

const (
	// StorageModeShared indicates that the resource is stored in system memory
	// accessible to both the CPU and the GPU.
	StorageModeShared StorageMode = 0

	// StorageModeManaged indicates that the resource exists as a synchronized
	// memory pair with one copy stored in system memory accessible to the CPU
	// and another copy stored in video memory accessible to the GPU.
	StorageModeManaged StorageMode = 1

	// StorageModePrivate indicates that the resource is stored in memory
	// only accessible to the GPU. In iOS and tvOS, the resource is stored in
	// system memory. In macOS, the resource is stored in video memory.
	StorageModePrivate StorageMode = 2

	// StorageModeMemoryless indicates that the resource is stored in on-tile memory,
	// without CPU or GPU memory backing. The contents of the on-tile memory are undefined
	// and do not persist; the only way to populate the resource is to render into it.
	// Memoryless resources are limited to temporary render targets (i.e., Textures configured
	// with a TextureDescriptor and used with a RenderPassAttachmentDescriptor).
	StorageModeMemoryless StorageMode = 3
)

// PixelFormat defines data formats that describe the organization
// and characteristics of individual pixels in a texture.
// https://developer.apple.com/documentation/metal/mtlpixelformat.
type PixelFormat uint8

// The data formats that describe the organization and characteristics
// of individual pixels in a texture.
const (
	PixelFormatRGBA8UNorm     PixelFormat = 70 // Ordinary format with four 8-bit normalized unsigned integer components in RGBA order.
	PixelFormatBGRA8UNorm     PixelFormat = 80 // Ordinary format with four 8-bit normalized unsigned integer components in BGRA order.
	PixelFormatBGRA8UNormSRGB PixelFormat = 81 // Ordinary format with four 8-bit normalized unsigned integer components in BGRA order with conversion between sRGB and linear space.
)

// TextureDescriptor configures new Texture objects.
// https://developer.apple.com/documentation/metal/mtltexturedescriptor.
type TextureDescriptor struct {
	PixelFormat PixelFormat
	Width       int
	Height      int
	StorageMode StorageMode
}

// Texture is a memory allocation for storing formatted
// image data that is accessible to the GPU.
// https://developer.apple.com/documentation/metal/mtltexture.
type Texture struct {
	texture unsafe.Pointer

	// width is the width of the texture image for the base level mipmap, in pixels.
	width int
	// height is the height of the texture image for the base level mipmap, in pixels.
	height int
}

// NewTexture returns a Texture that wraps an existing id<MTLTexture> pointer.
func NewTexture(texture unsafe.Pointer) Texture {
	return Texture{texture: texture}
}

// Release frees the current texture.
func (t Texture) Release() {
	C.Texture_Release(t.texture)
}

// ReplaceRegion copies a block of pixels into a section of texture slice 0.
// https://developer.apple.com/documentation/metal/mtltexture/1515464-replaceregion.
func (t Texture) ReplaceRegion(region Region, level int, pixelBytes []byte, bytesPerRow uintptr) {
	r := C.struct_Region{
		Origin: C.struct_Origin{
			X: C.uint_t(region.Origin.X),
			Y: C.uint_t(region.Origin.Y),
			Z: C.uint_t(region.Origin.Z),
		},
		Size: C.struct_Size{
			Width:  C.uint_t(region.Size.Width),
			Height: C.uint_t(region.Size.Height),
			Depth:  C.uint_t(region.Size.Depth),
		},
	}
	C.Texture_ReplaceRegion(t.texture, r, C.uint_t(level), unsafe.Pointer(&pixelBytes[0]), C.size_t(bytesPerRow))
}

// CommandQueue is a queue that organizes the order
// in which command buffers are executed by the GPU.
// https://developer.apple.com/documentation/metal/mtlcommandqueue.
type CommandQueue struct {
	commandQueue unsafe.Pointer
}

// MakeCommandBuffer creates a command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandqueue/1508686-makecommandbuffer.
func (cq CommandQueue) MakeCommandBuffer() CommandBuffer {
	return CommandBuffer{C.CommandQueue_MakeCommandBuffer(cq.commandQueue)}
}

// Release frees the command queue.
func (cq CommandQueue) Release() {
	C.CommandQueue_Release(cq.commandQueue)
}

// Drawable is a displayable resource that can be rendered or written to.
// https://developer.apple.com/documentation/metal/mtldrawable.
type Drawable interface {
	// Drawable returns the underlying id<MTLDrawable> pointer.
	Drawable() unsafe.Pointer
}

// CommandBuffer is a container that stores encoded commands
// that are committed to and executed by the GPU.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer.
type CommandBuffer struct {
	commandBuffer unsafe.Pointer
}

// PresentDrawable registers a drawable presentation to occur as soon as possible.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443029-presentdrawable.
func (cb CommandBuffer) PresentDrawable(d Drawable) {
	C.CommandBuffer_PresentDrawable(cb.commandBuffer, d.Drawable())
}

// Commit commits this command buffer for execution as soon as possible.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443003-commit.
func (cb CommandBuffer) Commit() {
	C.CommandBuffer_Commit(cb.commandBuffer)
}

// WaitUntilCompleted waits for the execution of this command buffer to complete.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443039-waituntilcompleted.
func (cb CommandBuffer) WaitUntilCompleted() {
	C.CommandBuffer_WaitUntilCompleted(cb.commandBuffer)
}

var commandBufferCompletedHandlers = sync.Map{} // map[unsafe.Pointer]func(){}

// AddCompletedHandler registers a block of code that Metal calls immediately after the GPU finishes executing the commands in the command buffer.
//
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1442997-addcompletedhandler
func (cb CommandBuffer) AddCompletedHandler(f func()) {
	commandBufferCompletedHandlers.Store(cb.commandBuffer, f)
	C.CommandBuffer_AddCompletedHandler(cb.commandBuffer)
}

//export commandBufferCompletedCallback
func commandBufferCompletedCallback(commandBuffer unsafe.Pointer) {
	f, ok := commandBufferCompletedHandlers.LoadAndDelete(commandBuffer)
	if !ok {
		return
	}

	f.(func())()
}

// Release frees the command buffer.
func (cb CommandBuffer) Release() {
	C.CommandBuffer_Release(cb.commandBuffer)
}

// MakeBlitCommandEncoder creates an encoder object that can encode
// memory operation (blit) commands into this command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443001-makeblitcommandencoder.
func (cb CommandBuffer) MakeBlitCommandEncoder() BlitCommandEncoder {
	return BlitCommandEncoder{CommandEncoder{C.CommandBuffer_MakeBlitCommandEncoder(cb.commandBuffer)}}
}

// ComputeCommandEncoder is for encoding commands in a compute pass.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder.
type ComputeCommandEncoder struct {
	CommandEncoder
}

// MakeComputeCommandEncoder creates an encoder object that can encode
// memory operation (blit) commands into this command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443001-makeblitcommandencoder.
func (cb CommandBuffer) MakeComputeCommandEncoder() ComputeCommandEncoder {
	return ComputeCommandEncoder{CommandEncoder{C.CommandBuffer_MakeComputeCommandEncoder(cb.commandBuffer)}}
}

// SetComputePipelineState sets the current compute pipeline state object.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443140-setcomputepipelinestate.
func (cce ComputeCommandEncoder) SetComputePipelineState(cps ComputePipelineState) {
	C.ComputeCommandEncoder_SetComputePipelineState(cce.commandEncoder, cps.computePipelineState)
}

// SetBytes sets a block of data for the compute shader.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443159-setbytes?language=objc.
func (cce ComputeCommandEncoder) SetBytes(b []byte, index int) {
	C.ComputeCommandEncoder_SetBytes(cce.commandEncoder, unsafe.Pointer(&b[0]), C.int(len(b)), C.int(index))
}

// SetBuffer sets a buffer for the compute function.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443126-setbuffer?language=objc
func (cce ComputeCommandEncoder) SetBuffer(b Buffer, offset, index int) {
	C.ComputeCommandEncoder_SetBuffer(cce.commandEncoder, b.buffer, C.int(offset), C.int(index))
}

func (cce ComputeCommandEncoder) DispatchThreads(threadsPerGrid, threadsPerThreadgroup Size) {
	C.ComputeCommandEncoder_DispatchThreads(cce.commandEncoder, C.struct_Size{
		Width:  C.uint_t(threadsPerGrid.Width),
		Height: C.uint_t(threadsPerGrid.Height),
		Depth:  C.uint_t(threadsPerGrid.Depth),
	}, C.struct_Size{
		Width:  C.uint_t(threadsPerThreadgroup.Width),
		Height: C.uint_t(threadsPerThreadgroup.Height),
		Depth:  C.uint_t(threadsPerThreadgroup.Depth),
	})
}

// CommandEncoder is an encoder that writes sequential GPU commands
// into a command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandencoder.
type CommandEncoder struct {
	commandEncoder unsafe.Pointer
}

// EndEncoding declares that all command generation from this encoder is completed.
// https://developer.apple.com/documentation/metal/mtlcommandencoder/1458038-endencoding.
func (ce CommandEncoder) EndEncoding() {
	C.CommandEncoder_EndEncoding(ce.commandEncoder)
}

// BlitCommandEncoder is an encoder that specifies resource copy
// and resource synchronization commands.
// https://developer.apple.com/documentation/metal/mtlblitcommandencoder.
type BlitCommandEncoder struct {
	CommandEncoder
}

// CopyFromTexture encodes a command to copy image data from a slice of
// a source texture into a slice of a destination texture.
// https://developer.apple.com/documentation/metal/mtlblitcommandencoder/1400754-copyfromtexture.
func (bce BlitCommandEncoder) CopyFromTexture(
	src Texture, srcSlice, srcLevel int, srcOrigin Origin, srcSize Size,
	dst Texture, dstSlice, dstLevel int, dstOrigin Origin,
) {
	C.BlitCommandEncoder_CopyFromTexture(
		bce.commandEncoder,
		src.texture, C.uint_t(srcSlice), C.uint_t(srcLevel),
		C.struct_Origin{
			X: C.uint_t(srcOrigin.X),
			Y: C.uint_t(srcOrigin.Y),
			Z: C.uint_t(srcOrigin.Z),
		},
		C.struct_Size{
			Width:  C.uint_t(srcSize.Width),
			Height: C.uint_t(srcSize.Height),
			Depth:  C.uint_t(srcSize.Depth),
		},
		dst.texture, C.uint_t(dstSlice), C.uint_t(dstLevel),
		C.struct_Origin{
			X: C.uint_t(dstOrigin.X),
			Y: C.uint_t(dstOrigin.Y),
			Z: C.uint_t(dstOrigin.Z),
		},
	)
}

// Release frees the blit command encoder.
func (bce BlitCommandEncoder) Release() {
	C.BlitCommandEncoder_Release(bce.commandEncoder)
}

// ResourceOptions defines optional arguments used to create
// and influence behavior of buffer and texture objects.
//
// https://developer.apple.com/documentation/metal/mtlresourceoptions.
type ResourceOptions uint16

const (
	resourceCPUCacheModeShift       = 0
	resourceStorageModeShift        = 4
	resourceHazardTrackingModeShift = 8
)

// CPUCacheMode is the CPU cache mode that defines the CPU mapping of a resource.
//
// https://developer.apple.com/documentation/metal/mtlcpucachemode.
type CPUCacheMode uint8

const (
	// CPUCacheModeDefaultCache is the default CPU cache mode for the resource.
	// Guarantees that read and write operations are executed in the expected order.
	CPUCacheModeDefaultCache CPUCacheMode = 0

	// CPUCacheModeWriteCombined is a write-combined CPU cache mode for the resource.
	// Optimized for resources that the CPU will write into, but never read.
	CPUCacheModeWriteCombined CPUCacheMode = 1
)

const (
	// ResourceCPUCacheModeDefaultCache is the default CPU cache mode for the resource.
	// Guarantees that read and write operations are executed in the expected order.
	ResourceCPUCacheModeDefaultCache ResourceOptions = ResourceOptions(CPUCacheModeDefaultCache) << resourceCPUCacheModeShift

	// ResourceCPUCacheModeWriteCombined is a write-combined CPU cache mode for the resource.
	// Optimized for resources that the CPU will write into, but never read.
	ResourceCPUCacheModeWriteCombined ResourceOptions = ResourceOptions(CPUCacheModeWriteCombined) << resourceCPUCacheModeShift

	// ResourceStorageModeShared indicates that the resource is stored in system memory
	// accessible to both the CPU and the GPU.
	ResourceStorageModeShared ResourceOptions = ResourceOptions(StorageModeShared) << resourceStorageModeShift

	// ResourceStorageModeManaged indicates that the resource exists as a synchronized
	// memory pair with one copy stored in system memory accessible to the CPU
	// and another copy stored in video memory accessible to the GPU.
	ResourceStorageModeManaged ResourceOptions = ResourceOptions(StorageModeManaged) << resourceStorageModeShift

	// ResourceStorageModePrivate indicates that the resource is stored in memory
	// only accessible to the GPU. In iOS and tvOS, the resource is stored
	// in system memory. In macOS, the resource is stored in video memory.
	ResourceStorageModePrivate ResourceOptions = ResourceOptions(StorageModePrivate) << resourceStorageModeShift

	// ResourceStorageModeMemoryless indicates that the resource is stored in on-tile memory,
	// without CPU or GPU memory backing. The contents of the on-tile memory are undefined
	// and do not persist; the only way to populate the resource is to render into it.
	// Memoryless resources are limited to temporary render targets (i.e., Textures configured
	// with a TextureDescriptor and used with a RenderPassAttachmentDescriptor).
	ResourceStorageModeMemoryless ResourceOptions = ResourceOptions(StorageModeMemoryless) << resourceStorageModeShift

	// ResourceHazardTrackingModeUntracked indicates that the command encoder dependencies
	// for this resource are tracked manually with Fence objects. This value is always set
	// for resources sub-allocated from a Heap object and may optionally be specified for
	// non-heap resources.
	ResourceHazardTrackingModeUntracked ResourceOptions = 1 << resourceHazardTrackingModeShift
)

// Buffer is a memory allocation for storing unformatted data
// that is accessible to the GPU.
//
// https://developer.apple.com/documentation/metal/mtlbuffer.
type Buffer struct {
	buffer unsafe.Pointer
}

func (b Buffer) Content() unsafe.Pointer {
	return C.Buffer_Content(b.buffer)
}

func (b Buffer) Release() {
	C.Buffer_Release(b.buffer)
}

// MakeBuffer allocates a new buffer of a given length
// and initializes its contents by copying existing data into it.
//
// The given bytes could be nil.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433429-makebuffer.
func (d Device) MakeBuffer(bytes unsafe.Pointer, length uintptr, opt ResourceOptions) Buffer {
	return Buffer{C.Device_MakeBuffer(d.device, bytes, C.size_t(length), C.uint16_t(opt))}
}

// CompileOptions specifies optional compilation settings for
// the graphics or compute functions within a library.
//
// https://developer.apple.com/documentation/metal/mtlcompileoptions.
type CompileOptions struct {
	LanguageVersion LanguageVersion

	// TODO: more options.
}

// https://developer.apple.com/documentation/metal/mtllanguageversion
type LanguageVersion int

const (
	LanguageVersion1_0 LanguageVersion = (1 << 16)
	LanguageVersion1_1 LanguageVersion = (1 << 16) + 1
	LanguageVersion1_2 LanguageVersion = (1 << 16) + 2
	LanguageVersion2_0 LanguageVersion = (2 << 16)
	LanguageVersion2_1 LanguageVersion = (2 << 16) + 1
	LanguageVersion2_2 LanguageVersion = (2 << 16) + 2
	LanguageVersion2_3 LanguageVersion = (2 << 16) + 3
	LanguageVersion2_4 LanguageVersion = (2 << 16) + 4
)

// Library is a collection of compiled graphics or compute functions.
//
// https://developer.apple.com/documentation/metal/mtllibrary.
type Library struct {
	library unsafe.Pointer
}

// MakeLibrary creates a new library that contains
// the functions stored in the specified source string.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433431-makelibrary.
func (d Device) MakeLibrary(source string, opt CompileOptions) (Library, error) {
	src := C.CString(source)
	defer C.free(unsafe.Pointer(src))

	copt := C.struct_CompileOption{
		languageVersion: C.uint_t(opt.LanguageVersion),
	}

	l := C.Device_MakeLibrary(d.device, src, copt) // TODO: opt.
	if l.Library == nil {
		return Library{}, errors.New(C.GoString(l.Error))
	}

	return Library{l.Library}, nil
}

// Function represents a programmable graphics or compute function executed by the GPU.
//
// https://developer.apple.com/documentation/metal/mtlfunction.
type Function struct {
	function unsafe.Pointer
}

// MakeFunction returns a pre-compiled, non-specialized function.
//
// https://developer.apple.com/documentation/metal/mtllibrary/1515524-makefunction.
func (l Library) MakeFunction(name string) (Function, error) {
	f := C.Library_MakeFunction(l.library, C.CString(name))
	if f == nil {
		return Function{}, fmt.Errorf("function %q not found", name)
	}

	return Function{f}, nil
}

// ComputePipelineState contains a compiled compute pipeline.
//
// https://developer.apple.com/documentation/metal/mtlcomputepipelinestate.
type ComputePipelineState struct {
	computePipelineState unsafe.Pointer
}

// MakeComputePipelineState creates a compute pipeline state object.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433427-newcomputepipelinestatewithfunct.
func (d Device) MakeComputePipelineState(fn Function) (ComputePipelineState, error) {
	cps := C.Device_MakeComputePipelineState(d.device, fn.function)
	if cps.ComputePipelineState == nil {
		return ComputePipelineState{}, errors.New(C.GoString(cps.Error))
	}

	return ComputePipelineState{cps.ComputePipelineState}, nil
}

func (cps ComputePipelineState) ThreadExecutionWidth() int {
	return int(C.ComputePipelineState_ThreadExecutionWidth(cps.computePipelineState))
}

func (cps ComputePipelineState) MaxTotalThreadsPerThreadgroup() int {
	return int(C.ComputePipelineState_MaxTotalThreadsPerThreadgroup(cps.computePipelineState))
}
