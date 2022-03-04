// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

// Package app provides the ability to run a window.
//
// A minimum window instance must implements Window interface that require
// a Size method to provide the size of the window. If the window instance
// additionaly implements a Draw method that returns an image.RGBA pointer,
// the created window will draw the provided image on the window.
//
// To handle resize, keyboard, and mouse events, the window instance must
// additionally implements OnResize, OnKey and OnMouse methods. See
// ResizeHandler, KeyboardHandler, and MouseHandler interfaces.
//
// Platform Specific Dependencies
//
// - Darwin: xcode-install --select
// - Linux: sudo apt install -y libx11-dev libgles2-mesa-dev libegl1-mesa-dev
package app
