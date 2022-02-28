// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

@import Metal;

#include "mtl.h"

// commandBufferCompletedCallback is an exported function from Go.
void commandBufferCompletedCallback(void *commandBuffer);

struct Device CreateSystemDefaultDevice() {
	id<MTLDevice> device = MTLCreateSystemDefaultDevice();
	if (!device) {
		struct Device d;
		d.Device = NULL;
		return d;
	}

	struct Device d;
	d.Device = device;
	d.Headless = device.headless;
	d.LowPower = device.lowPower;
	d.Removable = device.removable;
	d.RegistryID = device.registryID;
	d.Name = device.name.UTF8String;
	return d;
}

void * Device_MakeCommandQueue(void * device) {
	return [(id<MTLDevice>)device newCommandQueue];
}

void * Device_MakeTexture(void * device, struct TextureDescriptor descriptor) {
	MTLTextureDescriptor * textureDescriptor = [[MTLTextureDescriptor alloc] init];
	textureDescriptor.pixelFormat = descriptor.PixelFormat;
	textureDescriptor.width = descriptor.Width;
	textureDescriptor.height = descriptor.Height;
	textureDescriptor.storageMode = descriptor.StorageMode;
	return [(id<MTLDevice>)device newTextureWithDescriptor:textureDescriptor];
}

int MTLTexture_GetWidth(void *texture) {
	return ((id<MTLTexture>)texture).width;
}

int MTLTexture_GetHeight(void *texture) {
	return ((id<MTLTexture>)texture).height;
}

void Texture_ReplaceRegion(void * texture, struct Region region, uint_t level, void * pixelBytes, size_t bytesPerRow) {
	[(id<MTLTexture>)texture replaceRegion:(MTLRegion){{region.Origin.X, region.Origin.Y, region.Origin.Z}, {region.Size.Width, region.Size.Height, region.Size.Depth}}
	                           mipmapLevel:(NSUInteger)level
	                             withBytes:(void *)pixelBytes
	                           bytesPerRow:(NSUInteger)bytesPerRow];
}

void Texture_Release(void * texture) {
	[(id<MTLTexture>)texture release];
}

void * CommandQueue_MakeCommandBuffer(void * commandQueue) {
	return [(id<MTLCommandQueue>)commandQueue commandBuffer];
}

void CommandQueue_Release(void *commandQueue) {
  [(id<MTLCommandQueue>)commandQueue release];
}

void CommandEncoder_EndEncoding(void * commandEncoder) {
	[(id<MTLCommandEncoder>)commandEncoder endEncoding];
}

void * CommandBuffer_MakeBlitCommandEncoder(void * commandBuffer) {
	return [(id<MTLCommandBuffer>)commandBuffer blitCommandEncoder];
}

void BlitCommandEncoder_Release(void *blitCommandEncoder) {
  [(id<MTLBlitCommandEncoder>)blitCommandEncoder release];
}

void BlitCommandEncoder_CopyFromTexture(void * blitCommandEncoder,
	void * srcTexture, uint_t srcSlice, uint_t srcLevel, struct Origin srcOrigin, struct Size srcSize,
	void * dstTexture, uint_t dstSlice, uint_t dstLevel, struct Origin dstOrigin) {
	[(id<MTLBlitCommandEncoder>)blitCommandEncoder copyFromTexture:(id<MTLTexture>)srcTexture
	                                                   sourceSlice:(NSUInteger)srcSlice
	                                                   sourceLevel:(NSUInteger)srcLevel
	                                                  sourceOrigin:(MTLOrigin){srcOrigin.X, srcOrigin.Y, srcOrigin.Z}
	                                                    sourceSize:(MTLSize){srcSize.Width, srcSize.Height, srcSize.Depth}
	                                                     toTexture:(id<MTLTexture>)dstTexture
	                                              destinationSlice:(NSUInteger)dstSlice
	                                              destinationLevel:(NSUInteger)dstLevel
	                                             destinationOrigin:(MTLOrigin){dstOrigin.X, dstOrigin.Y, dstOrigin.Z}];
}

void CommandBuffer_PresentDrawable(void * commandBuffer, void * drawable) {
	[(id<MTLCommandBuffer>)commandBuffer presentDrawable:(id<MTLDrawable>)drawable];
}

void CommandBuffer_Commit(void * commandBuffer) {
	[(id<MTLCommandBuffer>)commandBuffer commit];
}

void CommandBuffer_WaitUntilCompleted(void * commandBuffer) {
	[(id<MTLCommandBuffer>)commandBuffer waitUntilCompleted];
}

void CommandBuffer_AddCompletedHandler(void *commandBuffer) {
	[(id<MTLCommandBuffer>)commandBuffer addCompletedHandler:^(id<MTLCommandBuffer> cb) {
		commandBufferCompletedCallback(cb);
	}];
}

void CommandBuffer_Release(void *commandBuffer) {
  [(id<MTLCommandBuffer>)commandBuffer release];
}