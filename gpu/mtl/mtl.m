// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This file is inspired by https://dmitri.shuralyov.com/gpu/mtl.

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

void * CommandBuffer_MakeComputeCommandEncoder(void * commandBuffer) {
	return [(id<MTLCommandBuffer>)commandBuffer computeCommandEncoder];
}

void ComputeCommandEncoder_SetComputePipelineState(void * computeCommandEncoder, void * computePipelineState) {
	[(id<MTLComputeCommandEncoder>)computeCommandEncoder setComputePipelineState:(id<MTLComputePipelineState>)computePipelineState];
}

void ComputeCommandEncoder_SetBytes(void * computeCommandEncoder, void *bytes, int length, int index) {
	[(id<MTLComputeCommandEncoder>)computeCommandEncoder setBytes:bytes length:length atIndex:index];
}

void ComputeCommandEncoder_SetBuffer(void * computeCommandEncoder, void *buffer, int offset, int index) {
	[(id<MTLComputeCommandEncoder>)computeCommandEncoder setBuffer:buffer
		offset:offset
		atIndex:index];
}

void * Buffer_Content(void *buffer) {
	return ((id<MTLBuffer>)buffer).contents;
}

void Buffer_Release(void *buffer) {
	[(id<MTLBuffer>)buffer release];
}

void ComputeCommandEncoder_DispatchThreads(void * computeCommandEncoder, struct Size threadsPerGrid, struct Size threadsPerThreadgroup) {
	[(id<MTLComputeCommandEncoder>)computeCommandEncoder dispatchThreads:(MTLSize){threadsPerGrid.Width, threadsPerGrid.Height, threadsPerGrid.Depth}
		threadsPerThreadgroup:(MTLSize){threadsPerThreadgroup.Width, threadsPerThreadgroup.Height, threadsPerThreadgroup.Depth}];
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


void * Device_MakeBuffer(void * device, const void * bytes, size_t length, uint16_t options) {
	if (bytes == NULL) {
		return [(id<MTLDevice>)device newBufferWithLength:(NSUInteger)length
			options:(MTLResourceOptions)options];
	} else {
		return [(id<MTLDevice>)device newBufferWithBytes:(const void *)bytes
			length:(NSUInteger)length
			options:(MTLResourceOptions)options];
	}
}

struct Library Device_MakeLibrary(void * device, const char * source, struct CompileOption opt) {
	MTLCompileOptions *compileOptions = [MTLCompileOptions new];
	compileOptions.languageVersion = (NSUInteger)opt.languageVersion;

	NSError * error;
	id<MTLLibrary> library = [(id<MTLDevice>)device
		newLibraryWithSource:[NSString stringWithUTF8String:source]
		options:compileOptions
		error:&error];

	struct Library l;
	l.Library = library;
	if (!library) {
		l.Error = error.localizedDescription.UTF8String;
	}
	return l;
}

void * Library_MakeFunction(void * library, const char * name) {
	return [(id<MTLLibrary>)library newFunctionWithName:[NSString stringWithUTF8String:name]];
}

struct ComputePipelineState Device_MakeComputePipelineState(void * device, void *function) {
	NSError * error;
	id<MTLComputePipelineState> computePipelineState = [(id<MTLDevice>)device newComputePipelineStateWithFunction:(id<MTLFunction>)(function)
	error:&error];

	struct ComputePipelineState cps;
	cps.ComputePipelineState = computePipelineState;
	if (!computePipelineState) {
		cps.Error = error.localizedDescription.UTF8String;
	}
	return cps;
}

int ComputePipelineState_ThreadExecutionWidth(void *cps) {
	return ((id<MTLComputePipelineState>)(cps)).threadExecutionWidth;
}

int ComputePipelineState_MaxTotalThreadsPerThreadgroup(void *cps) {
	return ((id<MTLComputePipelineState>)(cps)).maxTotalThreadsPerThreadgroup;
}
