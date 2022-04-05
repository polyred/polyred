// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

#include <stdlib.h>
#include <stdbool.h>

typedef unsigned long uint_t;
typedef unsigned char uint8_t;
typedef unsigned short uint16_t;
typedef unsigned long long uint64_t;

struct Device {
	void *       Device;
	bool         Headless;
	bool         LowPower;
	bool         Removable;
	uint64_t     RegistryID;
	const char * Name;
};

struct TextureDescriptor {
	uint16_t PixelFormat;
	uint_t   Width;
	uint_t   Height;
	uint8_t  StorageMode;
};

struct Origin {
	uint_t X;
	uint_t Y;
	uint_t Z;
};

struct Size {
	uint_t Width;
	uint_t Height;
	uint_t Depth;
};

struct Region {
	struct Origin Origin;
	struct Size   Size;
};

struct Device CreateSystemDefaultDevice();

// CommandQueue
void * Device_MakeCommandQueue(void * device) ;
void * CommandQueue_MakeCommandBuffer(void * commandQueue);
void CommandQueue_Release(void *commandQueue);

// CommandEncoder, BlitCommandEncoder, ComputeCommandEncoder
void CommandEncoder_EndEncoding(void * commandEncoder);
void * CommandBuffer_MakeBlitCommandEncoder(void * commandBuffer);
void * CommandBuffer_MakeComputeCommandEncoder(void * commandBuffer);
void ComputeCommandEncoder_SetComputePipelineState(void * computeCommandEncoder, void * computePipelineState);
void ComputeCommandEncoder_SetBytes(void * computeCommandEncoder, void *bytes, int length, int index);
void ComputeCommandEncoder_SetBuffer(void * computeCommandEncoder, void *buffer, int offset, int index);
void ComputeCommandEncoder_DispatchThreads(void * computeCommandEncoder, struct Size threadsPerGrid, struct Size threadsPerThreadgroup);

void BlitCommandEncoder_CopyFromTexture(void * blitCommandEncoder,
	void * srcTexture, uint_t srcSlice, uint_t srcLevel, struct Origin srcOrigin, struct Size srcSize,
	void * dstTexture, uint_t dstSlice, uint_t dstLevel, struct Origin dstOrigin);
void BlitCommandEncoder_Release(void *blitCommandEncoder);

// CommandBuffer
void CommandBuffer_PresentDrawable(void * commandBuffer, void * drawable);
void CommandBuffer_WaitUntilCompleted(void * commandBuffer);
void CommandBuffer_Commit(void * commandBuffer);
void CommandBuffer_AddCompletedHandler(void *commandBuffer);
void CommandBuffer_Release(void *commandBuffer);

// MTLTexture
void * Device_MakeTexture(void * device, struct TextureDescriptor descriptor);
int MTLTexture_GetWidth(void *texture);
int MTLTexture_GetHeight(void *texture);
void Texture_ReplaceRegion(void * texture, struct Region region, uint_t level, void * pixelBytes, size_t bytesPerRow);
void Texture_Release(void * texture);

void * Device_MakeBuffer(void * device, const void * bytes, size_t length, uint16_t options);
void * Buffer_Content(void *buffer);
void Buffer_Release(void *buffer);

struct Library {
	void *       Library;
	const char * Error;
};
struct CompileOption {
	uint_t languageVersion;
};

struct Library Device_MakeLibrary(void * device, const char * source, struct CompileOption opt);
void * Library_MakeFunction(void * library, const char * name);

struct ComputePipelineState {
	void *       ComputePipelineState;
	const char * Error;
};

struct ComputePipelineState Device_MakeComputePipelineState(void * device, void * function);
int ComputePipelineState_ThreadExecutionWidth(void *cps);
int ComputePipelineState_MaxTotalThreadsPerThreadgroup(void *cps);