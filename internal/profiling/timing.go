// Copyright 2022 The Polyred Authors. All rights reserved.
// Use of this source code is governed by a GPLv3 license that
// can be found in the LICENSE file.

package profiling

import (
	"fmt"
	"io"
	"os"
	"time"
)

var w io.Writer = os.Stdout

// SetWriter sets the profiling writer.
func SetWriter(writer io.Writer) {
	w = writer
}

// Timed returns a function for printing out the time elapced.
func Timed(name string) func() {
	start := time.Now()
	return func() {
		fmt.Fprintf(w, "%s...%v\n", name, time.Since(start))
	}
}
