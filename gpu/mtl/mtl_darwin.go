// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file is inspired by https://dmitri.shuralyov.com/gpu/mtl.

// Package mtl provides the access to Apple's Metal API.
// https://developer.apple.com/documentation/metal
//
// This package requires macOS version 10.13 or newer.
//
// It is cgo-free: the Objective-C runtime and the Metal framework are
// reached through github.com/ebitengine/purego instead of an Objective-C
// bridge, so the package builds with CGO_ENABLED=0.
package mtl

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"
)

// mtlCreateSystemDefaultDevice is the single C entry point we need from the
// Metal framework; everything else is Objective-C messaging.
var mtlCreateSystemDefaultDevice func() uintptr

func init() {
	for _, fw := range []string{
		"/System/Library/Frameworks/Foundation.framework/Foundation",
		"/System/Library/Frameworks/CoreGraphics.framework/CoreGraphics",
		"/System/Library/Frameworks/Metal.framework/Metal",
	} {
		if _, err := purego.Dlopen(fw, purego.RTLD_GLOBAL|purego.RTLD_NOW); err != nil {
			// Leave mtlCreateSystemDefaultDevice nil; Available() reports false.
			return
		}
	}
	metal, err := purego.Dlopen("/System/Library/Frameworks/Metal.framework/Metal", purego.RTLD_GLOBAL|purego.RTLD_NOW)
	if err != nil {
		return
	}
	purego.RegisterLibFunc(&mtlCreateSystemDefaultDevice, metal, "MTLCreateSystemDefaultDevice")
}

// Cached selectors (objc.RegisterName grabs a global lock, so cache the hot
// ones once).
var (
	selName                  = objc.RegisterName("name")
	selIsLowPower            = objc.RegisterName("isLowPower")
	selIsHeadless            = objc.RegisterName("isHeadless")
	selIsRemovable           = objc.RegisterName("isRemovable")
	selRegistryID            = objc.RegisterName("registryID")
	selUTF8String            = objc.RegisterName("UTF8String")
	selLocalizedDescription  = objc.RegisterName("localizedDescription")
	selStringWithUTF8String  = objc.RegisterName("stringWithUTF8String:")
	selAlloc                 = objc.RegisterName("alloc")
	selInit                  = objc.RegisterName("init")
	selRelease               = objc.RegisterName("release")
	selNewCommandQueue       = objc.RegisterName("newCommandQueue")
	selCommandBuffer         = objc.RegisterName("commandBuffer")
	selComputeCommandEncoder = objc.RegisterName("computeCommandEncoder")
	selBlitCommandEncoder    = objc.RegisterName("blitCommandEncoder")
	selSetComputePipeline    = objc.RegisterName("setComputePipelineState:")
	selSetBytes              = objc.RegisterName("setBytes:length:atIndex:")
	selSetBuffer             = objc.RegisterName("setBuffer:offset:atIndex:")
	selDispatchThreads       = objc.RegisterName("dispatchThreads:threadsPerThreadgroup:")
	selEndEncoding           = objc.RegisterName("endEncoding")
	selCommit                = objc.RegisterName("commit")
	selWaitUntilCompleted    = objc.RegisterName("waitUntilCompleted")
	selPresentDrawable       = objc.RegisterName("presentDrawable:")
	selAddCompletedHandler   = objc.RegisterName("addCompletedHandler:")
	selNewBufferWithBytes    = objc.RegisterName("newBufferWithBytes:length:options:")
	selNewBufferWithLength   = objc.RegisterName("newBufferWithLength:options:")
	selContents              = objc.RegisterName("contents")
	selNewLibraryWithSource  = objc.RegisterName("newLibraryWithSource:options:error:")
	selNewFunctionWithName   = objc.RegisterName("newFunctionWithName:")
	selNewComputePipeline    = objc.RegisterName("newComputePipelineStateWithFunction:error:")
	selThreadExecutionWidth  = objc.RegisterName("threadExecutionWidth")
	selMaxTotalThreads       = objc.RegisterName("maxTotalThreadsPerThreadgroup")
	selNewTextureWithDesc    = objc.RegisterName("newTextureWithDescriptor:")
	selSetPixelFormat        = objc.RegisterName("setPixelFormat:")
	selSetWidth              = objc.RegisterName("setWidth:")
	selSetHeight             = objc.RegisterName("setHeight:")
	selSetStorageMode        = objc.RegisterName("setStorageMode:")
	selTexWidth              = objc.RegisterName("width")
	selTexHeight             = objc.RegisterName("height")
	selReplaceRegion         = objc.RegisterName("replaceRegion:mipmapLevel:withBytes:bytesPerRow:")
	selCopyFromTexture       = objc.RegisterName("copyFromTexture:sourceSlice:sourceLevel:sourceOrigin:sourceSize:toTexture:destinationSlice:destinationLevel:destinationOrigin:")
	selSetLanguageVersion    = objc.RegisterName("setLanguageVersion:")
)

// Metal ABI structs passed by value through objc_msgSend. NSUInteger is 8
// bytes on the 64-bit platforms we target.
type mtlSize struct{ width, height, depth uint64 }
type mtlOrigin struct{ x, y, z uint64 }
type mtlRegion struct {
	origin mtlOrigin
	size   mtlSize
}

func toID(p unsafe.Pointer) objc.ID { return objc.ID(uintptr(p)) }
func toPtr(id objc.ID) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&id))
}

func nsString(s string) objc.ID {
	return objc.ID(objc.GetClass("NSString")).Send(selStringWithUTF8String, s)
}

func nsErrorString(e objc.ID) string {
	if e == 0 {
		return "unknown error"
	}
	return objc.Send[string](e.Send(selLocalizedDescription), selUTF8String)
}

// Device is abstract representation of the GPU that
// serves as the primary interface for a Metal app.
// https://developer.apple.com/documentation/metal/mtldevice.
type Device struct {
	device objc.ID

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
	if mtlCreateSystemDefaultDevice == nil {
		return Device{}, errors.New("metal is not supported on this system")
	}
	d := objc.ID(mtlCreateSystemDefaultDevice())
	if d == 0 {
		return Device{}, errors.New("metal is not supported on this system")
	}
	return Device{
		device:     d,
		Headless:   objc.Send[bool](d, selIsHeadless),
		LowPower:   objc.Send[bool](d, selIsLowPower),
		Removable:  objc.Send[bool](d, selIsRemovable),
		RegistryID: objc.Send[uint64](d, selRegistryID),
		Name:       objc.Send[string](d.Send(selName), selUTF8String),
	}, nil
}

// Available returns true if the current macOS supports Metal.
func (d Device) Available() bool {
	return d.device != 0
}

// Device returns the underlying id<MTLDevice> pointer.
func (d Device) Device() unsafe.Pointer { return toPtr(d.device) }

// MakeCommandQueue creates a serial command submission queue.
// https://developer.apple.com/documentation/metal/mtldevice/1433388-makecommandqueue.
func (d Device) MakeCommandQueue() CommandQueue {
	return CommandQueue{d.device.Send(selNewCommandQueue)}
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

func (s Size) c() mtlSize     { return mtlSize{uint64(s.Width), uint64(s.Height), uint64(s.Depth)} }
func (o Origin) c() mtlOrigin { return mtlOrigin{uint64(o.X), uint64(o.Y), uint64(o.Z)} }

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
	desc := objc.ID(objc.GetClass("MTLTextureDescriptor")).Send(selAlloc).Send(selInit)
	desc.Send(selSetPixelFormat, uint64(td.PixelFormat))
	desc.Send(selSetWidth, uint64(td.Width))
	desc.Send(selSetHeight, uint64(td.Height))
	desc.Send(selSetStorageMode, uint64(td.StorageMode))
	if td.Usage != 0 {
		desc.Send(selSetUsage, uint64(td.Usage))
	}
	texture := d.device.Send(selNewTextureWithDesc, desc)
	desc.Send(selRelease)
	return Texture{
		texture: texture,
		width:   int(objc.Send[uint64](texture, selTexWidth)),
		height:  int(objc.Send[uint64](texture, selTexHeight)),
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
	Usage       TextureUsage
}

// Texture is a memory allocation for storing formatted
// image data that is accessible to the GPU.
// https://developer.apple.com/documentation/metal/mtltexture.
type Texture struct {
	texture objc.ID

	// width is the width of the texture image for the base level mipmap, in pixels.
	width int
	// height is the height of the texture image for the base level mipmap, in pixels.
	height int
}

// NewTexture returns a Texture that wraps an existing id<MTLTexture> pointer.
func NewTexture(texture unsafe.Pointer) Texture {
	return Texture{texture: toID(texture)}
}

// Release frees the current texture.
func (t Texture) Release() {
	t.texture.Send(selRelease)
}

// ReplaceRegion copies a block of pixels into a section of texture slice 0.
// https://developer.apple.com/documentation/metal/mtltexture/1515464-replaceregion.
func (t Texture) ReplaceRegion(region Region, level int, pixelBytes []byte, bytesPerRow uintptr) {
	r := mtlRegion{origin: region.Origin.c(), size: region.Size.c()}
	t.texture.Send(selReplaceRegion, r, uint64(level), unsafe.Pointer(&pixelBytes[0]), uint64(bytesPerRow))
}

// CommandQueue is a queue that organizes the order
// in which command buffers are executed by the GPU.
// https://developer.apple.com/documentation/metal/mtlcommandqueue.
type CommandQueue struct {
	commandQueue objc.ID
}

// MakeCommandBuffer creates a command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandqueue/1508686-makecommandbuffer.
func (cq CommandQueue) MakeCommandBuffer() CommandBuffer {
	return CommandBuffer{cq.commandQueue.Send(selCommandBuffer)}
}

// Release frees the command queue.
func (cq CommandQueue) Release() {
	cq.commandQueue.Send(selRelease)
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
	commandBuffer objc.ID
}

// PresentDrawable registers a drawable presentation to occur as soon as possible.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443029-presentdrawable.
func (cb CommandBuffer) PresentDrawable(d Drawable) {
	cb.commandBuffer.Send(selPresentDrawable, toID(d.Drawable()))
}

// Commit commits this command buffer for execution as soon as possible.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443003-commit.
func (cb CommandBuffer) Commit() {
	cb.commandBuffer.Send(selCommit)
}

// WaitUntilCompleted waits for the execution of this command buffer to complete.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443039-waituntilcompleted.
func (cb CommandBuffer) WaitUntilCompleted() {
	cb.commandBuffer.Send(selWaitUntilCompleted)
}

// AddCompletedHandler registers a block of code that Metal calls immediately
// after the GPU finishes executing the commands in the command buffer.
//
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1442997-addcompletedhandler
func (cb CommandBuffer) AddCompletedHandler(f func()) {
	block := objc.NewBlock(func(block objc.Block, _ objc.ID) {
		f()
	})
	cb.commandBuffer.Send(selAddCompletedHandler, block)
}

// Release frees the command buffer.
func (cb CommandBuffer) Release() {
	cb.commandBuffer.Send(selRelease)
}

// MakeBlitCommandEncoder creates an encoder object that can encode
// memory operation (blit) commands into this command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443001-makeblitcommandencoder.
func (cb CommandBuffer) MakeBlitCommandEncoder() BlitCommandEncoder {
	return BlitCommandEncoder{CommandEncoder{cb.commandBuffer.Send(selBlitCommandEncoder)}}
}

// ComputeCommandEncoder is for encoding commands in a compute pass.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder.
type ComputeCommandEncoder struct {
	CommandEncoder
}

// MakeComputeCommandEncoder creates an encoder object that can encode
// compute commands into this command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandbuffer/1443009-makecomputecommandencoder.
func (cb CommandBuffer) MakeComputeCommandEncoder() ComputeCommandEncoder {
	return ComputeCommandEncoder{CommandEncoder{cb.commandBuffer.Send(selComputeCommandEncoder)}}
}

// SetComputePipelineState sets the current compute pipeline state object.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443140-setcomputepipelinestate.
func (cce ComputeCommandEncoder) SetComputePipelineState(cps ComputePipelineState) {
	cce.commandEncoder.Send(selSetComputePipeline, cps.computePipelineState)
}

// SetBytes sets a block of data for the compute shader.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443159-setbytes?language=objc.
func (cce ComputeCommandEncoder) SetBytes(b []byte, index int) {
	cce.commandEncoder.Send(selSetBytes, unsafe.Pointer(&b[0]), uint64(len(b)), uint64(index))
}

// SetBuffer sets a buffer for the compute function.
//
// https://developer.apple.com/documentation/metal/mtlcomputecommandencoder/1443126-setbuffer?language=objc
func (cce ComputeCommandEncoder) SetBuffer(b Buffer, offset, index int) {
	cce.commandEncoder.Send(selSetBuffer, b.buffer, uint64(offset), uint64(index))
}

func (cce ComputeCommandEncoder) DispatchThreads(threadsPerGrid, threadsPerThreadgroup Size) {
	cce.commandEncoder.Send(selDispatchThreads, threadsPerGrid.c(), threadsPerThreadgroup.c())
}

// CommandEncoder is an encoder that writes sequential GPU commands
// into a command buffer.
// https://developer.apple.com/documentation/metal/mtlcommandencoder.
type CommandEncoder struct {
	commandEncoder objc.ID
}

// EndEncoding declares that all command generation from this encoder is completed.
// https://developer.apple.com/documentation/metal/mtlcommandencoder/1458038-endencoding.
func (ce CommandEncoder) EndEncoding() {
	ce.commandEncoder.Send(selEndEncoding)
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
	bce.commandEncoder.Send(selCopyFromTexture,
		src.texture, uint64(srcSlice), uint64(srcLevel), srcOrigin.c(), srcSize.c(),
		dst.texture, uint64(dstSlice), uint64(dstLevel), dstOrigin.c(),
	)
}

// Release frees the blit command encoder.
func (bce BlitCommandEncoder) Release() {
	bce.commandEncoder.Send(selRelease)
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
	buffer objc.ID
}

func (b Buffer) Content() unsafe.Pointer {
	return objc.Send[unsafe.Pointer](b.buffer, selContents)
}

func (b Buffer) Release() {
	b.buffer.Send(selRelease)
}

// MakeBuffer allocates a new buffer of a given length
// and initializes its contents by copying existing data into it.
//
// The given bytes could be nil.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433429-makebuffer.
func (d Device) MakeBuffer(bytes unsafe.Pointer, length uintptr, opt ResourceOptions) Buffer {
	if bytes == nil {
		return Buffer{d.device.Send(selNewBufferWithLength, uint64(length), uint64(opt))}
	}
	return Buffer{d.device.Send(selNewBufferWithBytes, bytes, uint64(length), uint64(opt))}
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
	library objc.ID
}

// MakeLibrary creates a new library that contains
// the functions stored in the specified source string.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433431-makelibrary.
func (d Device) MakeLibrary(source string, opt CompileOptions) (Library, error) {
	var options objc.ID
	if opt.LanguageVersion != 0 {
		options = objc.ID(objc.GetClass("MTLCompileOptions")).Send(selAlloc).Send(selInit)
		options.Send(selSetLanguageVersion, uint64(opt.LanguageVersion))
		defer options.Send(selRelease)
	}

	var err objc.ID
	lib := d.device.Send(selNewLibraryWithSource, nsString(source), options, unsafe.Pointer(&err))
	if lib == 0 {
		return Library{}, errors.New(nsErrorString(err))
	}
	return Library{lib}, nil
}

// Function represents a programmable graphics or compute function executed by the GPU.
//
// https://developer.apple.com/documentation/metal/mtlfunction.
type Function struct {
	function objc.ID
}

// MakeFunction returns a pre-compiled, non-specialized function.
//
// https://developer.apple.com/documentation/metal/mtllibrary/1515524-makefunction.
func (l Library) MakeFunction(name string) (Function, error) {
	f := l.library.Send(selNewFunctionWithName, nsString(name))
	if f == 0 {
		return Function{}, fmt.Errorf("function %q not found", name)
	}
	return Function{f}, nil
}

// ComputePipelineState contains a compiled compute pipeline.
//
// https://developer.apple.com/documentation/metal/mtlcomputepipelinestate.
type ComputePipelineState struct {
	computePipelineState objc.ID
}

// MakeComputePipelineState creates a compute pipeline state object.
//
// https://developer.apple.com/documentation/metal/mtldevice/1433427-newcomputepipelinestatewithfunct.
func (d Device) MakeComputePipelineState(fn Function) (ComputePipelineState, error) {
	var err objc.ID
	cps := d.device.Send(selNewComputePipeline, fn.function, unsafe.Pointer(&err))
	if cps == 0 {
		return ComputePipelineState{}, errors.New(nsErrorString(err))
	}
	return ComputePipelineState{cps}, nil
}

func (cps ComputePipelineState) ThreadExecutionWidth() int {
	return int(objc.Send[uint64](cps.computePipelineState, selThreadExecutionWidth))
}

func (cps ComputePipelineState) MaxTotalThreadsPerThreadgroup() int {
	return int(objc.Send[uint64](cps.computePipelineState, selMaxTotalThreads))
}
