// Copyright 2021 Changkun Ou <changkun.de>. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package utils

import (
	"fmt"
	"time"
)

// Timed returns a function for printing out the time elapced.
func Timed(name string) func() {
	start := time.Now()
	return func() {
		fmt.Printf("%s...%v\n", name, time.Since(start))
	}
}
