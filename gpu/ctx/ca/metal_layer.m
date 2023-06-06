// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build darwin

@import AppKit;
@import Cocoa;
@import QuartzCore;
@import Metal;
@import MetalKit;

#include "_cgo_export.h"

typedef unsigned long uint_t;
typedef unsigned short uint16_t;

uint16_t MetalLayer_PixelFormat(CFTypeRef metalLayer) {
	return ((__bridge CAMetalLayer *)metalLayer).pixelFormat;
}

void MetalLayer_SetDevice(CFTypeRef metalLayer, CFTypeRef device) {
	((__bridge CAMetalLayer *)metalLayer).device = (__bridge id<MTLDevice>)device;
}

const char * MetalLayer_SetPixelFormat(CFTypeRef metalLayer, uint16_t pixelFormat) {
	@try {
		((__bridge CAMetalLayer *)metalLayer).pixelFormat = (MTLPixelFormat)pixelFormat;
	}
	@catch (NSException * exception) {
		return exception.reason.UTF8String;
	}
	return NULL;
}

const char * MetalLayer_SetMaximumDrawableCount(CFTypeRef metalLayer, uint_t maximumDrawableCount) {
	if (@available(macOS 10.13.2, *)) {
		@try {
			((__bridge CAMetalLayer *)metalLayer).maximumDrawableCount = (NSUInteger)maximumDrawableCount;
		}
		@catch (NSException * exception) {
			return exception.reason.UTF8String;
		}
	}
	return NULL;
}

void MetalLayer_SetDisplaySyncEnabled(CFTypeRef metalLayer, bool displaySyncEnabled) {
	((__bridge CAMetalLayer *)metalLayer).displaySyncEnabled = displaySyncEnabled;
}

void MetalLayer_SetDrawableSize(CFTypeRef metalLayer, double width, double height) {
	((__bridge CAMetalLayer *)metalLayer).drawableSize = (CGSize){width, height};
}

CFTypeRef MetalLayer_NextDrawable(CFTypeRef metalLayer) {
	return (__bridge CFTypeRef)([(__bridge CAMetalLayer *)metalLayer nextDrawable]);
}

CFTypeRef MetalDrawable_Texture(CFTypeRef metalDrawable) {
	return (__bridge CFTypeRef)(((__bridge id<CAMetalDrawable>)metalDrawable).texture);
}
