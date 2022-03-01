// Copyright 2022 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

//go:build !windows

package term

import (
	"os"
	"syscall"
	"unsafe"
)

func GetSize() (int, int, error) {
	var size [4]uint16
	if _, _, err := syscall.Syscall(syscall.SYS_IOCTL, uintptr(os.Stdout.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&size))); err != 0 {
		return 0, 0, err
	}
	rows, cols := size[0], size[1]

	return int(cols), int(rows), nil
}
