// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build windows

package term

import (
	"syscall"
	"unsafe"
)

type (
	short     int16
	word      uint16
	smallRect struct {
		left, top, right, bottom short
	}
	coord struct {
		x, y short
	}
)

var (
	kernel32                   = syscall.NewLazyDLL("kernel32.dll")
	getControlScreenBufferInfo = kernel32.NewProc("GetConsoleScreenBufferInfo")
)

type termScreenBufferInfo struct {
	size       coord
	cursorPos  coord
	attributes word
	window     smallRect
	maxWinSize coord
}

func GetSize() (int, int, error) {
	h, err := syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	if err != nil {
		return 0, 0, err
	}

	info := &termScreenBufferInfo{}
	r0, _, e1 := syscall.Syscall(getControlScreenBufferInfo.Addr(), 2, uintptr(h), uintptr(unsafe.Pointer(info)), 0)
	if r0 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	if err != nil {
		return 0, 0, err
	}

	return int(info.size.x), int(info.size.y), nil
}
