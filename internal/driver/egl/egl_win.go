// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build windows

package egl

import (
	"syscall"
	"unsafe"

	"poly.red/internal/driver/windows"
)

func NewDisplay() NativeDisplayType {
	hInst, err := windows.GetModuleHandle()
	if err != nil {
		panic(err)
	}

	wcls := windows.WndClassEx{
		CbSize:        uint32(unsafe.Sizeof(windows.WndClassEx{})),
		Style:         windows.CS_HREDRAW | windows.CS_VREDRAW | windows.CS_OWNDC,
		HInstance:     hInst,
		LpszClassName: syscall.StringToUTF16Ptr("polyred"),
	}
	cls, err := windows.RegisterClassEx(&wcls)
	if err != nil {
		panic(err)
	}

	hwnd, err := windows.CreateWindowEx(uint32(windows.WS_EX_APPWINDOW|windows.WS_EX_WINDOWEDGE),
		cls,
		"",
		windows.WS_OVERLAPPEDWINDOW|windows.WS_CLIPSIBLINGS|windows.WS_CLIPCHILDREN,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		windows.CW_USEDEFAULT, windows.CW_USEDEFAULT,
		0,
		0,
		hInst,
		0)
	if err != nil {
		panic(err)
	}
	hdc, err := windows.GetDC(hwnd)
	if err != nil {
		panic(err)
	}

	return NativeDisplayType(hdc)
}
