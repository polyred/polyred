// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// +build darwin

typedef unsigned long uint_t;
typedef unsigned short uint16_t;

void * Window_ContentView(void * window);
void View_SetLayer(void * view, void * layer);
void View_SetWantsLayer(void * view, bool wantsLayer);

void * MakeMetalLayer();

uint16_t     MetalLayer_PixelFormat(void * metalLayer);
void         MetalLayer_SetDevice(void * metalLayer, void * device);
const char * MetalLayer_SetPixelFormat(void * metalLayer, uint16_t pixelFormat);
const char * MetalLayer_SetMaximumDrawableCount(void * metalLayer, uint_t maximumDrawableCount);
void         MetalLayer_SetDisplaySyncEnabled(void * metalLayer, bool displaySyncEnabled);
void         MetalLayer_SetDrawableSize(void * metalLayer, double width, double height);
void *       MetalLayer_NextDrawable(void * metalLayer);

void * MetalDrawable_Texture(void * drawable);