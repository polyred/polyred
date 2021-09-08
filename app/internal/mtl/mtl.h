// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

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
void * CommandQueue_MakeCommandBuffer(void * commandQueue);
void CommandEncoder_EndEncoding(void * commandEncoder);
void * CommandBuffer_MakeBlitCommandEncoder(void * commandBuffer);
void BlitCommandEncoder_CopyFromTexture(void * blitCommandEncoder,
	void * srcTexture, uint_t srcSlice, uint_t srcLevel, struct Origin srcOrigin, struct Size srcSize,
	void * dstTexture, uint_t dstSlice, uint_t dstLevel, struct Origin dstOrigin);
void CommandBuffer_PresentDrawable(void * commandBuffer, void * drawable);
void CommandBuffer_WaitUntilCompleted(void * commandBuffer);
void CommandBuffer_Commit(void * commandBuffer);
void CommandBuffer_AddCompletedHandler(void *commandBuffer);
void * Device_MakeTexture(void * device, struct TextureDescriptor descriptor);
int MTLTexture_GetWidth(void *texture);
int MTLTexture_GetHeight(void *texture);
void Texture_ReplaceRegion(void * texture, struct Region region, uint_t level, void * pixelBytes, size_t bytesPerRow);
void * Device_MakeCommandQueue(void * device) ;