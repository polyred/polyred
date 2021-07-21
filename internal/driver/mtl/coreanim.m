// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// +build darwin

#import <QuartzCore/QuartzCore.h>
#import <Cocoa/Cocoa.h>
#import "coreanim.h"

typedef unsigned long uint_t;
typedef unsigned short uint16_t;


void * Window_ContentView(void * window) {
	return ((NSWindow *)window).contentView;
}

void View_SetLayer(void * view, void * layer) {
	((NSView *)view).layer = (CALayer *)layer;
}

void View_SetWantsLayer(void * view, bool wantsLayer) {
	((NSView *)view).wantsLayer = wantsLayer;
}

void * MakeMetalLayer() {
	return [[CAMetalLayer alloc] init];
}

uint16_t MetalLayer_PixelFormat(void * metalLayer) {
	return ((CAMetalLayer *)metalLayer).pixelFormat;
}

void MetalLayer_SetDevice(void * metalLayer, void * device) {
	((CAMetalLayer *)metalLayer).device = (id<MTLDevice>)device;
}

const char * MetalLayer_SetPixelFormat(void * metalLayer, uint16_t pixelFormat) {
	@try {
		((CAMetalLayer *)metalLayer).pixelFormat = (MTLPixelFormat)pixelFormat;
	}
	@catch (NSException * exception) {
		return exception.reason.UTF8String;
	}
	return NULL;
}

const char * MetalLayer_SetMaximumDrawableCount(void * metalLayer, uint_t maximumDrawableCount) {
	if (@available(macOS 10.13.2, *)) {
		@try {
			((CAMetalLayer *)metalLayer).maximumDrawableCount = (NSUInteger)maximumDrawableCount;
		}
		@catch (NSException * exception) {
			return exception.reason.UTF8String;
		}
	}
	return NULL;
}

void MetalLayer_SetDisplaySyncEnabled(void * metalLayer, bool displaySyncEnabled) {
	((CAMetalLayer *)metalLayer).displaySyncEnabled = displaySyncEnabled;
}

void MetalLayer_SetDrawableSize(void * metalLayer, double width, double height) {
	((CAMetalLayer *)metalLayer).drawableSize = (CGSize){width, height};
}

void * MetalLayer_NextDrawable(void * metalLayer) {
	return [(CAMetalLayer *)metalLayer nextDrawable];
}

void * MetalDrawable_Texture(void * metalDrawable) {
	return ((id<CAMetalDrawable>)metalDrawable).texture;
}
