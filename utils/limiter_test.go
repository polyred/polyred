// Copyright 2021 Changkun Ou. All rights reserved.
// Use of this source code is governed by a license
// that can be found in the LICENSE file.

package utils_test

import (
	"runtime"
	"testing"

	"changkun.de/x/ddd/utils"
)

var (
	f = func() {
		p := 0
		for i := 0; i < 100000; i++ {
			p += i
		}
	}
)

func BenchmarkLimiter(b *testing.B) {
	np := runtime.GOMAXPROCS(0)
	l := utils.NewLimiter(np)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Execute(f)
	}
	l.Wait()
}
