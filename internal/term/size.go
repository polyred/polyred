// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !windows
// +build !windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

// termSize returns the terminal columns, rows, and cursor aspect ratio
func GetSize() (int, int, error) {
	var size [4]uint16
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size)), 0, 0, 0); err != 0 {
		return 0, 0, err
	}

	return int(size[0]), int(size[1]), nil

	// const defaultRatio = 16.0 / 9
	// rows, cols, width, height := size[0], size[1], size[2], size[3]
	// var whratio = defaultRatio
	// if width > 0 && height > 0 {
	// 	whratio = float64(height/rows) / float64(width/cols)
	// }
}
