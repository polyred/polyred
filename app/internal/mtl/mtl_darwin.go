// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package mtl provides the absolute minimum access to
// Apple's Metal API for drawing images on a CAMetalLayer.
// https://developer.apple.com/documentation/metal
//
// This package requires macOS version 10.13 or newer.
package mtl

/*
#cgo CFLAGS: -Werror -fmodules -x objective-c
#cgo LDFLAGS: -framework Metal
#include "mtl.h"
*/
import "C"
import (
	"errors"
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
func (t Texture) ReplaceRegion(region Region, level int, pixelBytes *byte, bytesPerRow uintptr) {
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
	C.Texture_ReplaceRegion(t.texture, r, C.uint_t(level), unsafe.Pointer(pixelBytes), C.size_t(bytesPerRow))
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
// Reference: https://developer.apple.com/documentation/metal/mtlcommandbuffer/1442997-addcompletedhandler
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
